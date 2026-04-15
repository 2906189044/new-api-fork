/*
Copyright (C) 2025 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/

export function formatSubscriptionDuration(plan, t) {
  const unit = plan?.duration_unit || 'month';
  const value = plan?.duration_value || 1;
  const unitLabels = {
    year: t('年'),
    month: t('个月'),
    day: t('天'),
    hour: t('小时'),
    custom: t('自定义'),
  };
  if (unit === 'custom') {
    const seconds = plan?.custom_seconds || 0;
    if (seconds >= 86400) return `${Math.floor(seconds / 86400)} ${t('天')}`;
    if (seconds >= 3600) return `${Math.floor(seconds / 3600)} ${t('小时')}`;
    return `${seconds} ${t('秒')}`;
  }
  return `${value} ${unitLabels[unit] || unit}`;
}

export function formatSubscriptionResetPeriod(plan, t) {
  const period = plan?.quota_reset_period || 'never';
  if (period === 'never') return t('不重置');
  if (period === 'daily') return t('每天');
  if (period === 'weekly') return t('每周');
  if (period === 'monthly') return t('每月');
  if (period === 'custom') {
    const seconds = Number(plan?.quota_reset_custom_seconds || 0);
    if (seconds >= 86400) return `${Math.floor(seconds / 86400)} ${t('天')}`;
    if (seconds >= 3600) return `${Math.floor(seconds / 3600)} ${t('小时')}`;
    if (seconds >= 60) return `${Math.floor(seconds / 60)} ${t('分钟')}`;
    return `${seconds} ${t('秒')}`;
  }
  return t('不重置');
}

export function getSubscriptionQuotaLabel(plan, t) {
  const period = plan?.quota_reset_period || 'never';
  if (period === 'daily') return t('日限额');
  if (period === 'weekly') return t('周限额');
  if (period === 'monthly') return t('月限额');
  return t('总额度');
}

export function getSubscriptionSummaryQuotaLabel(plan, t) {
  const period = plan?.quota_reset_period || 'never';
  if (period === 'daily') return t('今日额度');
  return getSubscriptionQuotaLabel(plan, t);
}

export function getSubscriptionSummaryQuotaHint(plan, t) {
  const period = plan?.quota_reset_period || 'never';
  if (period === 'daily') return t('额度每日刷新');
  return '';
}

export function formatSubscriptionPrice(plan) {
  const amount = Number(plan?.price_amount || 0);
  const currency = String(plan?.currency || 'CNY').toUpperCase();
  const formattedAmount = amount.toFixed(Number.isInteger(amount) ? 0 : 2);
  if (currency === 'USD') return `$${formattedAmount}`;
  if (currency === 'CNY') return `¥${formattedAmount}`;
  return `${currency} ${formattedAmount}`;
}

export function getSubscriptionDisplayConfig(plan) {
  const raw = plan?.display_config;
  if (!raw || typeof raw !== 'string') {
    return {
      feature_points: [],
    };
  }
  try {
    const parsed = JSON.parse(raw);
    return {
      display_title: parsed?.display_title || '',
      display_subtitle: parsed?.display_subtitle || '',
      eyebrow: parsed?.eyebrow || '',
      highlight_tag: parsed?.highlight_tag || '',
      cta_text: parsed?.cta_text || '',
      price_caption: parsed?.price_caption || '',
      quota_caption: parsed?.quota_caption || '',
      theme: parsed?.theme || 'default',
      is_featured: parsed?.is_featured === true,
      feature_points: Array.isArray(parsed?.feature_points)
        ? parsed.feature_points.filter(
            (item) => typeof item === 'string' && item.trim() !== '',
          )
        : [],
    };
  } catch {
    return {
      feature_points: [],
    };
  }
}

export function getSubscriptionFeatureSlots(points = [], slotCount = 5) {
  const normalized = Array.isArray(points)
    ? points
        .filter((item) => typeof item === 'string')
        .map((item) => item.trim())
        .slice(0, slotCount)
    : [];
  while (normalized.length < slotCount) {
    normalized.push('');
  }
  return normalized;
}

export function getSubscriptionBenefitSlots(items = [], slotCount = 4) {
  const normalized = Array.isArray(items)
    ? items
        .filter((item) => item && typeof item.label === 'string')
        .slice(0, slotCount)
    : [];
  while (normalized.length < slotCount) {
    normalized.push(null);
  }
  return normalized;
}

export function getSubscriptionCardLayoutConfig() {
  return {
    cardPadding: 'p-4.5 md:p-5',
    cardMinHeight: 'min-h-[476px]',
    headerMinHeight: 'min-h-[92px]',
    eyebrowSlotHeight: 'h-[20px]',
    subtitleMinHeight: 64,
    featureBlockMinHeight: 'min-h-[144px]',
    benefitBlockMinHeight: 'min-h-[120px]',
    ctaPaddingTop: 'pt-4',
  };
}

export function shouldClampSubscriptionCopy(isMobile) {
  return !isMobile;
}

export function resolveSubscriptionDisplayTitle(plan, config = {}) {
  return config?.display_title || plan?.title || '';
}

export function resolveSubscriptionDisplaySubtitle(plan, config = {}) {
  return config?.display_subtitle || plan?.subtitle || '';
}
