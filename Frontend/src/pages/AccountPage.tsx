import React, { useState, useEffect } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { Header } from '../components/Header';
import { Footer } from '../components/Footer';
import { Toast } from '../components/Toast';
import { Countdown } from '../components/Countdown';
import { useAuthStore } from '../store/authStore';
import { ordersService } from '../services/ordersService';
import { itemsService } from '../services/itemsService';
import { apiFetch } from '../services/api';
import { useCartStore } from '../store/cartStore';
import { useToastStore } from '../store/toastStore';
import { formatCurrency } from '../utils/format';
import type { Order, Item } from '../types';

interface UserBid {
  id: string;
  itemId: string;
  bidder: string;
  amount: number;
  status: string;
  timestamp: string;
  itemTitle: string;
  itemLotNumber: string;
  currentBid: number;
  endTime?: string;
}

export const AccountPage: React.FC = () => {
  const user = useAuthStore((state) => state.user);
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated);
  const signOut = useAuthStore((state) => state.signOut);
  const addItemToCart = useCartStore((state) => state.addItem);
  const addToast = useToastStore((state) => state.addToast);
  const navigate = useNavigate();

  const [orders, setOrders] = useState<Order[]>([]);
  const [userBids, setUserBids] = useState<UserBid[]>([]);
  const [wonItems, setWonItems] = useState<Item[]>([]);
  const [myListings, setMyListings] = useState<Item[]>([]);
  const [loadingBids, setLoadingBids] = useState(true);
  const [loadingWon, setLoadingWon] = useState(true);
  const [loadingListings, setLoadingListings] = useState(true);

  useEffect(() => {
    if (!isAuthenticated) {
      navigate('/auth');
      return;
    }

    // Load orders
    ordersService.getOrders().then(setOrders).catch(() => setOrders([]));

    // Load user's active bids from real API
    setLoadingBids(true);
    apiFetch<UserBid[]>('/account/bids')
      .then((bids) => setUserBids(bids))
      .catch(() => setUserBids([]))
      .finally(() => setLoadingBids(false));

    // Load won items from real API
    setLoadingWon(true);
    apiFetch<Item[]>('/account/won')
      .then((items) => setWonItems(items))
      .catch(() => setWonItems([]))
      .finally(() => setLoadingWon(false));

    // Load user's created listings
    setLoadingListings(true);
    itemsService
      .getMyListings()
      .then((items) => setMyListings(items))
      .catch(() => setMyListings([]))
      .finally(() => setLoadingListings(false));
  }, [isAuthenticated, navigate]);

  // Compute active bids: only bids on items that haven't ended yet
  const activeBids = userBids.filter((b) => {
    if (b.status !== 'active') return false;
    if (b.endTime && new Date(b.endTime).getTime() < Date.now()) return false;
    return true;
  });

  // Highest bid per item (deduplicate)
  const highestBidByItem = activeBids.reduce<Record<string, UserBid>>((acc, bid) => {
    if (!acc[bid.itemId] || bid.amount > acc[bid.itemId].amount) {
      acc[bid.itemId] = bid;
    }
    return acc;
  }, {});
  const uniqueActiveBids = Object.values(highestBidByItem);

  const handleClaimWonItem = (item: Item) => {
    addItemToCart(item.id, 1);
    addToast(`Added won lot "${item.title}" to cart for settlement`, 'success');
    navigate('/cart');
  };

  if (!isAuthenticated) return null;

  return (
    <div className="min-h-screen flex flex-col bg-background text-primary">
      <Header />

      <main className="flex-1 max-w-[1440px] w-full mx-auto px-3 py-6 text-left">
        {/* User Profile Bar */}
        <div className="flex flex-col sm:flex-row sm:items-center justify-between hairline-b pb-4 mb-6 gap-2">
          <div>
            <h1 className="text-lg font-semibold text-primary">{user?.name}</h1>
            <p className="text-xs text-secondary font-mono">
              Account: {user?.id} • Verified Collector ({user?.email})
            </p>
          </div>
          <div className="flex items-center gap-2 self-start sm:self-auto">
            <Link
              to="/sell"
              className="btn-primary text-xs py-1 px-3 whitespace-nowrap"
            >
              + List Item for Sale / Auction
            </Link>
            <button
              type="button"
              onClick={() => {
                signOut();
                addToast('Signed out', 'info');
                navigate('/auth');
              }}
              className="btn-secondary text-xs py-1 px-3"
            >
              Sign Out
            </button>
          </div>
        </div>

        {/* 3 Main Sections */}
        <div className="space-y-8">
          {/* SECTION 1: ACTIVE BIDS */}
          <section>
            <div className="flex items-center justify-between hairline-b pb-2 mb-3">
              <h2 className="text-xs font-semibold uppercase tracking-wider text-secondary">
                1. Active Auction Bids ({loadingBids ? '…' : uniqueActiveBids.length})
              </h2>
              {uniqueActiveBids.length > 0 && (
                <span className="text-xs text-success font-medium">● Live</span>
              )}
            </div>

            {loadingBids ? (
              <div className="p-4 border border-hairline text-xs text-secondary bg-surface/20 animate-pulse">
                Loading your bids…
              </div>
            ) : uniqueActiveBids.length === 0 ? (
              <div className="p-4 border border-hairline text-xs text-secondary bg-surface/20">
                You have no active bids.{' '}
                <Link to="/items?type=auction" className="underline">
                  Browse auctions →
                </Link>
              </div>
            ) : (
              <div className="border border-hairline overflow-x-auto">
                <table className="w-full text-sm border-collapse">
                  <thead>
                    <tr className="bg-surface hairline-b text-xs text-secondary text-left">
                      <th className="py-2 px-3">Lot</th>
                      <th className="py-2 px-3">Item Description</th>
                      <th className="py-2 px-3">Your Bid</th>
                      <th className="py-2 px-3">Status</th>
                      <th className="py-2 px-3">Ends</th>
                      <th className="py-2 px-3 text-right">Action</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-hairline">
                    {uniqueActiveBids.map((bid) => {
                      const isHighest = bid.amount >= bid.currentBid;
                      return (
                        <tr key={bid.id} className="hover:bg-surface/30">
                          <td className="py-2 px-3 font-mono text-xs">{bid.itemLotNumber || bid.itemId}</td>
                          <td className="py-2 px-3">
                            <Link
                              to={`/items/${bid.itemId}`}
                              className="font-semibold text-primary hover:underline block truncate max-w-sm"
                            >
                              {bid.itemTitle || bid.itemId}
                            </Link>
                            <span className="text-xs text-secondary">
                              Current: {formatCurrency(bid.currentBid)}
                            </span>
                          </td>
                          <td className="py-2 px-3 font-mono font-semibold">
                            {formatCurrency(bid.amount)}
                          </td>
                          <td className="py-2 px-3">
                            {isHighest ? (
                              <span className="text-xs text-success font-semibold">● Highest</span>
                            ) : (
                              <span className="text-xs text-accent font-semibold">● Outbid</span>
                            )}
                          </td>
                          <td className="py-2 px-3 text-xs text-secondary">
                            {bid.endTime ? (
                              <Countdown endTime={bid.endTime} />
                            ) : (
                              'Open'
                            )}
                          </td>
                          <td className="py-2 px-3 text-right">
                            <Link
                              to={`/items/${bid.itemId}#bid`}
                              className="btn-secondary text-xs py-0.5 px-2"
                            >
                              {isHighest ? 'View Lot' : 'Increase Bid'}
                            </Link>
                          </td>
                        </tr>
                      );
                    })}
                  </tbody>
                </table>
              </div>
            )}
          </section>

          {/* SECTION 2: WON ITEMS READY FOR SETTLEMENT */}
          <section>
            <div className="flex items-center justify-between hairline-b pb-2 mb-3">
              <h2 className="text-xs font-semibold uppercase tracking-wider text-secondary">
                2. Won Auction Lots Awaiting Payment ({loadingWon ? '…' : wonItems.length})
              </h2>
              <span className="text-xs text-secondary">Ready for invoice settlement</span>
            </div>

            {loadingWon ? (
              <div className="p-4 border border-hairline text-xs text-secondary bg-surface/20 animate-pulse">
                Loading won items…
              </div>
            ) : wonItems.length === 0 ? (
              <div className="p-4 border border-hairline text-xs text-secondary bg-surface/20">
                No won lots awaiting settlement yet.
              </div>
            ) : (
              <div className="border border-hairline divide-y divide-hairline">
                {wonItems.map((item) => (
                  <div
                    key={item.id}
                    className="p-3 bg-surface/20 flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3"
                  >
                    <div className="flex items-center gap-3">
                      <img
                        src={item.images[0]}
                        alt={item.title}
                        className="w-12 h-12 object-contain p-0.5 border border-hairline bg-surface flex-shrink-0"
                      />
                      <div>
                        <div className="text-xs font-mono text-secondary">LOT #{item.lotNumber} [WINNER]</div>
                        <Link
                          to={`/items/${item.id}`}
                          className="font-semibold text-primary hover:underline text-sm"
                        >
                          {item.title}
                        </Link>
                        <div className="text-xs text-success font-medium">
                          Hammer Price: {formatCurrency(item.currentBid ?? item.price ?? 0)}
                        </div>
                      </div>
                    </div>

                    <button
                      type="button"
                      onClick={() => handleClaimWonItem(item)}
                      className="btn-primary text-xs py-1.5 px-3 self-end sm:self-auto"
                    >
                      Add to Cart for Payment →
                    </button>
                  </div>
                ))}
              </div>
            )}
          </section>

          {/* SECTION 3: ORDER HISTORY */}
          <section>
            <div className="flex items-center justify-between hairline-b pb-2 mb-3">
              <h2 className="text-xs font-semibold uppercase tracking-wider text-secondary">
                3. Historical Orders &amp; Paid Invoices ({orders.length})
              </h2>
            </div>

            {orders.length === 0 ? (
              <div className="p-4 border border-hairline text-xs text-secondary bg-surface/20">
                No completed invoices in your purchase record yet.
              </div>
            ) : (
              <div className="border border-hairline divide-y divide-hairline">
                {orders.map((ord) => (
                  <div key={ord.id} className="p-3 flex items-center justify-between text-xs">
                    <div>
                      <span className="font-semibold font-mono text-primary mr-3">{ord.id}</span>
                      <span className="text-secondary">
                        {new Date(ord.createdAt).toLocaleDateString()}
                      </span>
                      <span className="text-secondary ml-2">
                        • {ord.items.length} line items
                      </span>
                    </div>
                    <div className="flex items-center gap-3">
                      <span className="font-mono font-semibold">{formatCurrency(ord.total)}</span>
                      <Link to={`/order/${ord.id}`} className="hover:underline text-primary">
                        View Receipt
                      </Link>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </section>

          {/* SECTION 4: MY LISTED LOTS */}
          <section>
            <div className="flex items-center justify-between hairline-b pb-2 mb-3">
              <div className="flex items-center gap-2">
                <h2 className="text-xs font-semibold uppercase tracking-wider text-secondary">
                  4. My Consigned / Listed Lots ({loadingListings ? '…' : myListings.length})
                </h2>
                <span className="text-xs text-secondary font-mono">Active on public catalog</span>
              </div>
              <Link
                to="/sell"
                className="btn-primary text-xs py-1 px-3"
              >
                + List New Item
              </Link>
            </div>

            {loadingListings ? (
              <div className="p-4 border border-hairline text-xs text-secondary bg-surface/20 animate-pulse">
                Loading your listings…
              </div>
            ) : myListings.length === 0 ? (
              <div className="p-6 border border-hairline text-center bg-surface/20 space-y-2">
                <p className="text-xs text-secondary">You have not listed any lots for auction or sale yet.</p>
                <Link to="/sell" className="btn-primary text-xs py-1.5 px-4 inline-block">
                  Put an Item Up for Auction or Sale →
                </Link>
              </div>
            ) : (
              <div className="border border-hairline divide-y divide-hairline">
                {myListings.map((it) => (
                  <div key={it.id} className="p-3 flex items-center justify-between text-xs hover:bg-surface/30 gap-4">
                    <div className="flex items-center gap-3 min-w-0">
                      {it.images && it.images.length > 0 && (
                        <img
                          src={it.images[0]}
                          alt={it.title}
                          className="w-10 h-10 object-contain p-0.5 border border-hairline bg-surface flex-shrink-0"
                        />
                      )}
                      <div className="min-w-0">
                        <span className="font-mono text-secondary mr-2">#{it.lotNumber}</span>
                        <Link to={`/items/${it.id}`} className="font-semibold text-primary hover:underline">
                          {it.title}
                        </Link>
                        <div className="text-[11px] text-secondary">
                          Format: <span className="uppercase font-mono">{it.purchaseMode}</span> • Category: {it.category}
                        </div>
                      </div>
                    </div>

                    <div className="flex items-center gap-4 flex-shrink-0">
                      <div className="text-right">
                        <div className="font-mono font-semibold text-primary">
                          {it.purchaseMode === 'auction' ? formatCurrency(it.currentBid ?? it.startingBid ?? 0) : formatCurrency(it.price ?? 0)}
                        </div>
                        <div className="text-[10px] text-secondary">
                          {it.purchaseMode === 'auction' ? `${it.bidCount || 0} bids placed` : `Stock: ${it.stock || 1}`}
                        </div>
                      </div>

                      <Link to={`/items/${it.id}`} className="btn-secondary text-xs py-0.5 px-2.5">
                        View Live Lot
                      </Link>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </section>
        </div>
      </main>

      <Footer />
      <Toast />
    </div>
  );
};
