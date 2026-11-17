import type { Order, ContactInfo, ShippingAddress, CartItem, Item } from '../types';
import { apiFetch } from './api';
import { mockItems } from '../mocks/items';
import { trackPurchase } from './analytics';

let orderCounter = 1000;
const orders: Order[] = [];

export const ordersService = {
  async placeOrder(
    cartItems: CartItem[],
    contact: ContactInfo,
    shippingAddress: ShippingAddress,
  ): Promise<Order & { paymentUrl?: string; pidx?: string }> {
    try {
      const resp = await apiFetch<{ order: Order; payment?: { paymentUrl?: string; pidx?: string } }>('/orders', {
        method: 'POST',
        body: JSON.stringify({
          contact,
          shippingAddress,
          cartItemIds: cartItems.map((c) => c.itemId),
        }),
      });

      const order = resp.order;
      trackPurchase(
        order.id,
        order.total,
        order.items.map((i) => ({
          id: i.item.id,
          title: i.item.title,
          price: i.item.purchaseMode === 'auction' ? (i.item.currentBid ?? 0) : (i.item.price ?? 0),
          quantity: i.quantity,
        }))
      );

      return {
        ...order,
        paymentUrl: resp.payment?.paymentUrl,
        pidx: resp.payment?.pidx,
      };
    } catch {
      // Offline fallback
      const resolvedItems: { item: Item; quantity: number }[] = [];
      let subtotal = 0;
      let shippingCost = 0;

      for (const ci of cartItems) {
        const item = mockItems.find((i) => i.id === ci.itemId);
        if (!item) continue;
        resolvedItems.push({ item, quantity: ci.quantity });
        const price = item.purchaseMode === 'auction' ? (item.currentBid ?? 0) : (item.price ?? 0);
        subtotal += price * ci.quantity;
        shippingCost += item.shipping.cost;
      }

      const buyersPremium = Math.round(subtotal * 0.1);
      const total = subtotal + shippingCost + buyersPremium;

      const order: Order = {
        id: `ORD-${++orderCounter}`,
        items: resolvedItems,
        contact,
        shippingAddress,
        subtotal,
        shippingCost,
        buyersPremium,
        total,
        createdAt: new Date().toISOString(),
      };

      orders.push(order);
      trackPurchase(
        order.id,
        order.total,
        order.items.map((i) => ({
          id: i.item.id,
          title: i.item.title,
          price: i.item.purchaseMode === 'auction' ? (i.item.currentBid ?? 0) : (i.item.price ?? 0),
          quantity: i.quantity,
        }))
      );

      return order;
    }
  },

  async getOrders(): Promise<Order[]> {
    try {
      return await apiFetch<Order[]>('/orders');
    } catch {
      return [...orders];
    }
  },

  async getOrderById(id: string): Promise<Order | null> {
    try {
      return await apiFetch<Order>(`/orders/${id}`);
    } catch {
      return orders.find((o) => o.id === id) ?? null;
    }
  },

  async verifyPayment(pidx: string): Promise<{ status: string; transactionId: string; amount: number; message: string }> {
    return await apiFetch('/payments/verify', {
      method: 'POST',
      body: JSON.stringify({ pidx }),
    });
  },
};
