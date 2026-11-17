import type { BidHistoryEntry } from '../types';
import { apiFetch, getAuthToken } from './api';
import { trackPlaceBid } from './analytics';
import { formatCurrency } from '../utils/format';

export const bidsService = {
  async placeBid(itemId: string, amount: number): Promise<{ success: boolean; message: string }> {
    // Strictly require auth — never fall back to mock if backend is reachable
    const token = getAuthToken();
    if (!token || token === 'mock-jwt-token') {
      // Allow mock-jwt-token only as a development fallback; block no-token entirely
      if (!token) {
        return { success: false, message: 'You must be signed in to place a bid.' };
      }
    }

    try {
      const resp = await apiFetch<{ success: boolean; message: string; bid?: any }>(`/items/${itemId}/bids`, {
        method: 'POST',
        body: JSON.stringify({ amount }),
      });
      trackPlaceBid(itemId, `Item ${itemId}`, amount);
      return { success: resp.success, message: resp.message };
    } catch (err: any) {
      // Always surface real backend errors (401, 400, etc.) to the user.
      // Never silently succeed with a mock write when the backend is running.
      const msg = err?.message || 'Failed to place bid. Please try again.';
      return { success: false, message: msg };
    }
  },

  async getBidHistory(itemId: string): Promise<BidHistoryEntry[]> {
    try {
      return await apiFetch<BidHistoryEntry[]>(`/items/${itemId}/bids`);
    } catch {
      return [];
    }
  },

  async cancelBid(bidId: string): Promise<{ success: boolean; message: string }> {
    try {
      await apiFetch(`/bids/${bidId}`, { method: 'DELETE' });
      return { success: true, message: 'Bid cancelled successfully.' };
    } catch (err: any) {
      return { success: false, message: err.message || 'Failed to cancel bid.' };
    }
  },

  getMinimumBid(item: { currentBid?: number; startingBid?: number; minimumIncrement?: number; bidCount?: number }): number {
    const current = item.currentBid ?? item.startingBid ?? 0;
    const increment = item.minimumIncrement ?? 100;
    // If no bids yet, minimum is startingBid itself; otherwise current + increment
    if (!item.bidCount || item.bidCount === 0) {
      return item.startingBid ?? 0;
    }
    return current + increment;
  },

  formatMinBidLabel(item: { currentBid?: number; startingBid?: number; minimumIncrement?: number; bidCount?: number }): string {
    const minBid = bidsService.getMinimumBid(item);
    return formatCurrency(minBid);
  },
};
