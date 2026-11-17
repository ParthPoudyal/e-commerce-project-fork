declare global {
  interface Window {
    dataLayer: any[];
    gtag?: (...args: any[]) => void;
  }
}

const MEASUREMENT_ID = import.meta.env.VITE_GA4_MEASUREMENT_ID;

export function initGA4(): void {
  if (!MEASUREMENT_ID) {
    return;
  }

  // Inject Google Analytics script tag
  const script = document.createElement('script');
  script.async = true;
  script.src = `https://www.googletagmanager.com/gtag/js?id=${MEASUREMENT_ID}`;
  document.head.appendChild(script);

  window.dataLayer = window.dataLayer || [];
  window.gtag = function () {
    window.dataLayer.push(arguments);
  };
  window.gtag('js', new Date());
  window.gtag('config', MEASUREMENT_ID, { send_page_view: true });
}

export function trackEvent(eventName: string, params?: Record<string, any>): void {
  if (typeof window !== 'undefined' && window.gtag && MEASUREMENT_ID) {
    window.gtag('event', eventName, params);
  }
}

export function trackViewItem(item: { id: string; title: string; category?: string; price?: number }): void {
  trackEvent('view_item', {
    currency: 'NPR',
    value: item.price ?? 0,
    items: [
      {
        item_id: item.id,
        item_name: item.title,
        item_category: item.category,
        price: item.price ?? 0,
      },
    ],
  });
}

export function trackAddToCart(item: { id: string; title: string; category?: string; price?: number; quantity?: number }): void {
  trackEvent('add_to_cart', {
    currency: 'NPR',
    value: (item.price ?? 0) * (item.quantity ?? 1),
    items: [
      {
        item_id: item.id,
        item_name: item.title,
        item_category: item.category,
        price: item.price ?? 0,
        quantity: item.quantity ?? 1,
      },
    ],
  });
}

export function trackBeginCheckout(items: { id: string; title: string; price: number; quantity: number }[], total: number): void {
  trackEvent('begin_checkout', {
    currency: 'NPR',
    value: total,
    items: items.map((i) => ({
      item_id: i.id,
      item_name: i.title,
      price: i.price,
      quantity: i.quantity,
    })),
  });
}

export function trackPurchase(orderId: string, total: number, items: { id: string; title: string; price: number; quantity: number }[]): void {
  trackEvent('purchase', {
    transaction_id: orderId,
    currency: 'NPR',
    value: total,
    items: items.map((i) => ({
      item_id: i.id,
      item_name: i.title,
      price: i.price,
      quantity: i.quantity,
    })),
  });
}

export function trackPlaceBid(itemId: string, itemTitle: string, amount: number): void {
  trackEvent('place_bid', {
    currency: 'NPR',
    value: amount,
    item_id: itemId,
    item_name: itemTitle,
  });
}
