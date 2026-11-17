export function formatCurrency(val: number): string {
  return `NPR ${new Intl.NumberFormat('en-US', { maximumFractionDigits: 0 }).format(val)}`;
}
