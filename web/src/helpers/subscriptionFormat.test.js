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

import test from 'node:test';
import assert from 'node:assert/strict';
import {
  formatSubscriptionPrice,
  getSubscriptionBenefitSlots,
  getSubscriptionCardLayoutConfig,
  getSubscriptionFeatureSlots,
  getSubscriptionDisplayConfig,
  getSubscriptionQuotaLabel,
  getSubscriptionSummaryQuotaHint,
  getSubscriptionSummaryQuotaLabel,
  resolveSubscriptionDisplaySubtitle,
  resolveSubscriptionDisplayTitle,
  shouldClampSubscriptionCopy,
} from './subscriptionFormat.js';

const t = (value) => value;

test('getSubscriptionQuotaLabel maps reset periods to quota labels', () => {
  assert.equal(
    getSubscriptionQuotaLabel({ quota_reset_period: 'daily' }, t),
    '日限额',
  );
  assert.equal(
    getSubscriptionQuotaLabel({ quota_reset_period: 'weekly' }, t),
    '周限额',
  );
  assert.equal(
    getSubscriptionQuotaLabel({ quota_reset_period: 'monthly' }, t),
    '月限额',
  );
  assert.equal(
    getSubscriptionQuotaLabel({ quota_reset_period: 'never' }, t),
    '总额度',
  );
  assert.equal(getSubscriptionQuotaLabel({}, t), '总额度');
});

test('daily subscription summaries use today quota copy', () => {
  assert.equal(
    getSubscriptionSummaryQuotaLabel({ quota_reset_period: 'daily' }, t),
    '今日额度',
  );
  assert.equal(
    getSubscriptionSummaryQuotaHint({ quota_reset_period: 'daily' }, t),
    '额度每日刷新',
  );
  assert.equal(getSubscriptionSummaryQuotaLabel({}, t), '总额度');
  assert.equal(getSubscriptionSummaryQuotaHint({}, t), '');
});

test('formatSubscriptionPrice follows the plan currency instead of quota display settings', () => {
  assert.equal(
    formatSubscriptionPrice({ price_amount: 99, currency: 'CNY' }),
    '¥99',
  );
  assert.equal(
    formatSubscriptionPrice({ price_amount: 19.9, currency: 'USD' }),
    '$19.90',
  );
  assert.equal(formatSubscriptionPrice({ price_amount: 88 }), '¥88');
});

test('getSubscriptionDisplayConfig parses valid json and keeps known fields', () => {
  const config = getSubscriptionDisplayConfig({
    display_config:
      '{"display_title":"Plus","display_subtitle":"适合高频使用","feature_points":["高速响应","低延迟"]}',
  });

  assert.equal(config.display_title, 'Plus');
  assert.equal(config.display_subtitle, '适合高频使用');
  assert.deepEqual(config.feature_points, ['高速响应', '低延迟']);
});

test('subscription display title and subtitle fall back to logical fields', () => {
  const plan = {
    title: '基础套餐',
    subtitle: '默认副标题',
    display_config: '',
  };
  const config = getSubscriptionDisplayConfig(plan);

  assert.equal(resolveSubscriptionDisplayTitle(plan, config), '基础套餐');
  assert.equal(resolveSubscriptionDisplaySubtitle(plan, config), '默认副标题');
});

test('getSubscriptionBenefitSlots pads missing benefit rows to a fixed slot count', () => {
  const benefits = [{ label: '有效期: 1 个月' }, { label: '总额度: 450万' }];

  assert.deepEqual(getSubscriptionBenefitSlots(benefits, 4), [
    { label: '有效期: 1 个月' },
    { label: '总额度: 450万' },
    null,
    null,
  ]);
});

test('getSubscriptionFeatureSlots pads missing feature points to a fixed slot count', () => {
  assert.deepEqual(
    getSubscriptionFeatureSlots(['功能 A', '功能 B'], 5),
    ['功能 A', '功能 B', '', '', ''],
  );
  assert.deepEqual(
    getSubscriptionFeatureSlots(['A', 'B', 'C', 'D', 'E', 'F'], 5),
    ['A', 'B', 'C', 'D', 'E'],
  );
});

test('shouldClampSubscriptionCopy disables copy clamping on mobile layouts', () => {
  assert.equal(shouldClampSubscriptionCopy(true), false);
  assert.equal(shouldClampSubscriptionCopy(false), true);
});

test('getSubscriptionCardLayoutConfig exposes the tightened shared card rhythm', () => {
  assert.deepEqual(getSubscriptionCardLayoutConfig(), {
    cardPadding: 'p-4.5 md:p-5',
    cardMinHeight: 'min-h-[476px]',
    headerMinHeight: 'min-h-[92px]',
    eyebrowSlotHeight: 'h-[20px]',
    subtitleMinHeight: 64,
    featureBlockMinHeight: 'min-h-[144px]',
    benefitBlockMinHeight: 'min-h-[120px]',
    ctaPaddingTop: 'pt-4',
  });
});
