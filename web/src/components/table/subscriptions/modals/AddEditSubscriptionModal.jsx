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

import React, { useEffect, useState, useRef } from 'react';
import {
  Avatar,
  Button,
  Card,
  Col,
  Form,
  Row,
  Select,
  SideSheet,
  Space,
  Spin,
  Tag,
  Typography,
} from '@douyinfe/semi-ui';
import {
  IconCalendarClock,
  IconClose,
  IconCreditCard,
  IconSave,
} from '@douyinfe/semi-icons';
import { Clock, RefreshCw, Sparkles } from 'lucide-react';
import { API, showError, showSuccess } from '../../../../helpers';
import {
  quotaToDisplayAmount,
  displayAmountToQuota,
} from '../../../../helpers/quota';
import { useIsMobile } from '../../../../hooks/common/useIsMobile';
import {
  getSubscriptionDisplayConfig,
  getSubscriptionQuotaLabel,
} from '../../../../helpers/subscriptionFormat';

const { Text, Title } = Typography;

const durationUnitOptions = [
  { value: 'year', label: '年' },
  { value: 'month', label: '月' },
  { value: 'day', label: '日' },
  { value: 'hour', label: '小时' },
  { value: 'custom', label: '自定义(秒)' },
];

const resetPeriodOptions = [
  { value: 'never', label: '不重置' },
  { value: 'daily', label: '每天' },
  { value: 'weekly', label: '每周' },
  { value: 'monthly', label: '每月' },
  { value: 'custom', label: '自定义(秒)' },
];

const displayThemeOptions = [
  { value: 'default', label: '默认' },
  { value: 'featured', label: '推荐高亮' },
  { value: 'dark', label: '深色强调' },
];

const bonusModelScopeOptions = [
  { value: 'gpt_series', label: 'GPT 系列' },
];

const displayFeatureFields = [
  'feature_point_1',
  'feature_point_2',
  'feature_point_3',
  'feature_point_4',
  'feature_point_5',
];

const AddEditSubscriptionModal = ({
  visible,
  handleClose,
  editingPlan,
  placement = 'left',
  refresh,
  t,
}) => {
  const [loading, setLoading] = useState(false);
  const [groupOptions, setGroupOptions] = useState([]);
  const [groupLoading, setGroupLoading] = useState(false);
  const isMobile = useIsMobile();
  const formApiRef = useRef(null);
  const isEdit = editingPlan?.plan?.id !== undefined;
  const formKey = isEdit ? `edit-${editingPlan?.plan?.id}` : 'create';

  const getInitValues = () => ({
    title: '',
    subtitle: '',
    display_title: '',
    display_subtitle: '',
    display_eyebrow: '',
    display_highlight_tag: '',
    display_cta_text: '',
    display_price_caption: '',
    display_quota_caption: '',
    display_theme: 'default',
    display_featured: false,
    display_notes: '',
    feature_point_1: '',
    feature_point_2: '',
    feature_point_3: '',
    feature_point_4: '',
    feature_point_5: '',
    price_amount: 0,
    currency: 'CNY',
    duration_unit: 'month',
    duration_value: 1,
    custom_seconds: 0,
    quota_reset_period: 'never',
    quota_reset_custom_seconds: 0,
    enabled: true,
    visible_to_user: true,
    stackable_bonus: false,
    bonus_model_scope: '',
    sort_order: 0,
    max_purchase_per_user: 0,
    total_amount: 0,
    upgrade_group: '',
    stripe_price_id: '',
    creem_product_id: '',
  });

  const buildFormValues = () => {
    const base = getInitValues();
    if (editingPlan?.plan?.id === undefined) return base;
    const p = editingPlan.plan || {};
    const displayConfig = getSubscriptionDisplayConfig(p);
    const featurePoints = displayConfig.feature_points || [];
    return {
      ...base,
      title: p.title || '',
      subtitle: p.subtitle || '',
      display_title: displayConfig.display_title || '',
      display_subtitle: displayConfig.display_subtitle || '',
      display_eyebrow: displayConfig.eyebrow || '',
      display_highlight_tag: displayConfig.highlight_tag || '',
      display_cta_text: displayConfig.cta_text || '',
      display_price_caption: displayConfig.price_caption || '',
      display_quota_caption: displayConfig.quota_caption || '',
      display_theme: displayConfig.theme || 'default',
      display_featured: displayConfig.is_featured === true,
      display_notes: p.display_notes || '',
      feature_point_1: featurePoints[0] || '',
      feature_point_2: featurePoints[1] || '',
      feature_point_3: featurePoints[2] || '',
      feature_point_4: featurePoints[3] || '',
      feature_point_5: featurePoints[4] || '',
      price_amount: Number(p.price_amount || 0),
      currency: 'CNY',
      duration_unit: p.duration_unit || 'month',
      duration_value: Number(p.duration_value || 1),
      custom_seconds: Number(p.custom_seconds || 0),
      quota_reset_period: p.quota_reset_period || 'never',
      quota_reset_custom_seconds: Number(p.quota_reset_custom_seconds || 0),
      enabled: p.enabled !== false,
      visible_to_user: p.visible_to_user !== false,
      stackable_bonus: p.stackable_bonus === true,
      bonus_model_scope: p.bonus_model_scope || '',
      sort_order: Number(p.sort_order || 0),
      max_purchase_per_user: Number(p.max_purchase_per_user || 0),
      total_amount: Number(
        quotaToDisplayAmount(p.total_amount || 0).toFixed(2),
      ),
      upgrade_group: p.upgrade_group || '',
      stripe_price_id: p.stripe_price_id || '',
      creem_product_id: p.creem_product_id || '',
    };
  };

  useEffect(() => {
    if (!visible) return;
    setGroupLoading(true);
    API.get('/api/group')
      .then((res) => {
        if (res.data?.success) {
          setGroupOptions(res.data?.data || []);
        } else {
          setGroupOptions([]);
        }
      })
      .catch(() => setGroupOptions([]))
      .finally(() => setGroupLoading(false));
  }, [visible]);

  const submit = async (values) => {
    if (!values.title || values.title.trim() === '') {
      showError(t('套餐标题不能为空'));
      return;
    }
    setLoading(true);
    try {
      const featurePoints = displayFeatureFields
        .map((field) => String(values[field] || '').trim())
        .filter(Boolean);
      const displayConfig = {
        display_title: String(values.display_title || '').trim(),
        display_subtitle: String(values.display_subtitle || '').trim(),
        eyebrow: String(values.display_eyebrow || '').trim(),
        highlight_tag: String(values.display_highlight_tag || '').trim(),
        cta_text: String(values.display_cta_text || '').trim(),
        price_caption: String(values.display_price_caption || '').trim(),
        quota_caption: String(values.display_quota_caption || '').trim(),
        theme: values.display_theme || 'default',
        is_featured: values.display_featured === true,
        feature_points: featurePoints,
      };
      const payload = {
        plan: {
          ...values,
          display_config: JSON.stringify(displayConfig),
          display_notes: String(values.display_notes || '').trim(),
          price_amount: Number(values.price_amount || 0),
          currency: 'CNY',
          duration_value: Number(values.duration_value || 0),
          custom_seconds: Number(values.custom_seconds || 0),
          quota_reset_period: values.quota_reset_period || 'never',
          quota_reset_custom_seconds:
            values.quota_reset_period === 'custom'
              ? Number(values.quota_reset_custom_seconds || 0)
              : 0,
          sort_order: Number(values.sort_order || 0),
          max_purchase_per_user: Number(values.max_purchase_per_user || 0),
          total_amount: displayAmountToQuota(values.total_amount),
          upgrade_group: values.stackable_bonus ? '' : values.upgrade_group || '',
          stackable_bonus: values.stackable_bonus === true,
          bonus_model_scope: values.stackable_bonus
            ? values.bonus_model_scope || ''
            : '',
        },
      };
      if (editingPlan?.plan?.id) {
        const res = await API.put(
          `/api/subscription/admin/plans/${editingPlan.plan.id}`,
          payload,
        );
        if (res.data?.success) {
          showSuccess(t('更新成功'));
          handleClose();
          refresh?.();
        } else {
          showError(res.data?.message || t('更新失败'));
        }
      } else {
        const res = await API.post('/api/subscription/admin/plans', payload);
        if (res.data?.success) {
          showSuccess(t('创建成功'));
          handleClose();
          refresh?.();
        } else {
          showError(res.data?.message || t('创建失败'));
        }
      }
    } catch (e) {
      showError(t('请求失败'));
    } finally {
      setLoading(false);
    }
  };

  return (
    <>
      <SideSheet
        placement={placement}
        title={
          <Space>
            {isEdit ? (
              <Tag color='blue' shape='circle'>
                {t('更新')}
              </Tag>
            ) : (
              <Tag color='green' shape='circle'>
                {t('新建')}
              </Tag>
            )}
            <Title heading={4} className='m-0'>
              {isEdit ? t('更新套餐信息') : t('创建新的订阅套餐')}
            </Title>
          </Space>
        }
        bodyStyle={{ padding: '0' }}
        visible={visible}
        width={isMobile ? '100%' : 600}
        footer={
          <div className='flex justify-end bg-white'>
            <Space>
              <Button
                theme='solid'
                onClick={() => formApiRef.current?.submitForm()}
                icon={<IconSave />}
                loading={loading}
              >
                {t('提交')}
              </Button>
              <Button
                theme='light'
                type='primary'
                onClick={handleClose}
                icon={<IconClose />}
              >
                {t('取消')}
              </Button>
            </Space>
          </div>
        }
        closeIcon={null}
        onCancel={handleClose}
      >
        <Spin spinning={loading}>
          <Form
            key={formKey}
            initValues={buildFormValues()}
            getFormApi={(api) => (formApiRef.current = api)}
            onSubmit={submit}
          >
            {({ values }) => (
              <div className='p-2'>
                {/* 基本信息 */}
                <Card className='!rounded-2xl shadow-sm border-0 mb-4'>
                  <div className='flex items-center mb-2'>
                    <Avatar
                      size='small'
                      color='blue'
                      className='mr-2 shadow-md'
                    >
                      <IconCalendarClock size={16} />
                    </Avatar>
                    <div>
                      <Text className='text-lg font-medium'>
                        {t('基本信息')}
                      </Text>
                      <div className='text-xs text-gray-600'>
                        {t('套餐的基本信息和定价')}
                      </div>
                    </div>
                  </div>

                  <Row gutter={12}>
                    <Col span={24}>
                      <Form.Input
                        field='title'
                        label={t('套餐标题')}
                        placeholder={t('例如：基础套餐')}
                        required
                        rules={[
                          { required: true, message: t('请输入套餐标题') },
                        ]}
                        showClear
                      />
                    </Col>

                    <Col span={24}>
                      <Form.TextArea
                        field='subtitle'
                        label={t('默认副标题')}
                        placeholder={t('例如：适合轻度使用')}
                        autosize={{ minRows: 2, maxRows: 5 }}
                        extraText={t(
                          '未设置前端展示副标题时，将回退显示这里的内容',
                        )}
                        showClear
                      />
                    </Col>

                    <Col span={12}>
                      <Form.InputNumber
                        field='price_amount'
                        label={t('实付金额')}
                        prefix='¥'
                        required
                        min={0}
                        precision={2}
                        rules={[{ required: true, message: t('请输入金额') }]}
                        style={{ width: '100%' }}
                      />
                    </Col>

                    <Col span={12}>
                      <Form.InputNumber
                        field='total_amount'
                        label={getSubscriptionQuotaLabel(
                          { quota_reset_period: values.quota_reset_period },
                          t,
                        )}
                        required
                        min={0}
                        precision={2}
                        rules={[{ required: true, message: t('请输入总额度') }]}
                        extraText={`${t('0 表示不限')} · ${t('原生额度')}：${displayAmountToQuota(
                          values.total_amount,
                        )}`}
                        style={{ width: '100%' }}
                      />
                    </Col>

                    <Col span={12}>
                      <Form.Select
                        field='upgrade_group'
                        label={t('升级分组')}
                        showClear
                        loading={groupLoading}
                        placeholder={t('不升级')}
                        extraText={t(
                          '购买或手动新增订阅会升级到该分组；当套餐失效/过期或手动作废/删除后，将回退到升级前分组。回退不会立即生效，通常会有几分钟延迟。',
                        )}
                      >
                        <Select.Option value=''>{t('不升级')}</Select.Option>
                        {(groupOptions || []).map((g) => (
                          <Select.Option key={g} value={g}>
                            {g}
                          </Select.Option>
                        ))}
                      </Form.Select>
                    </Col>

                    <Col span={12}>
                      <Form.Input
                        field='currency'
                        label={t('币种')}
                        disabled
                        extraText={t('由全站货币展示设置统一控制')}
                      />
                    </Col>

                    <Col span={12}>
                      <Form.InputNumber
                        field='sort_order'
                        label={t('排序')}
                        precision={0}
                        style={{ width: '100%' }}
                      />
                    </Col>

                    <Col span={12}>
                      <Form.InputNumber
                        field='max_purchase_per_user'
                        label={t('购买上限')}
                        min={0}
                        precision={0}
                        extraText={t('0 表示不限')}
                        style={{ width: '100%' }}
                      />
                    </Col>

                    <Col span={12}>
                      <Form.Switch
                        field='enabled'
                        label={t('启用状态')}
                        size='large'
                      />
                    </Col>

                    <Col span={12}>
                      <Form.Switch
                        field='visible_to_user'
                        label={t('前台可见')}
                        extraText={t(
                          '关闭后普通用户不会在套餐页看到该套餐，但管理员仍可在用户订阅管理中手动开通',
                        )}
                        size='large'
                      />
                    </Col>

                    <Col span={12}>
                      <Form.Switch
                        field='stackable_bonus'
                        label={t('附加权益套餐')}
                        extraText={t(
                          '开启后不会覆盖用户主分组，可与其他套餐叠加使用',
                        )}
                        size='large'
                      />
                    </Col>

                    <Col span={12}>
                      <Form.Select
                        field='bonus_model_scope'
                        label={t('附加权益模型范围')}
                        disabled={!values.stackable_bonus}
                        placeholder={t('请选择模型范围')}
                        extraText={t(
                          '当前仅用于限制隐藏福利套餐可消耗的模型系列',
                        )}
                      >
                        {bonusModelScopeOptions.map((option) => (
                          <Select.Option
                            key={option.value}
                            value={option.value}
                          >
                            {option.label}
                          </Select.Option>
                        ))}
                      </Form.Select>
                    </Col>
                  </Row>
                </Card>

                <Card className='!rounded-2xl shadow-sm border-0 mb-4'>
                  <div className='flex items-center mb-2'>
                    <Avatar
                      size='small'
                      color='violet'
                      className='mr-2 shadow-md'
                    >
                      <Sparkles size={16} />
                    </Avatar>
                    <div>
                      <Text className='text-lg font-medium'>
                        {t('前端展示配置')}
                      </Text>
                      <div className='text-xs text-gray-600'>
                        {t('用于控制套餐卡片的商业化展示内容')}
                      </div>
                    </div>
                  </div>

                  <Row gutter={12}>
                    <Col span={12}>
                      <Form.Input
                        field='display_title'
                        label={t('展示标题')}
                        placeholder={t('例如：Plus')}
                        showClear
                      />
                    </Col>
                    <Col span={12}>
                      <Form.Input
                        field='display_eyebrow'
                        label={t('眉题')}
                        placeholder={t('例如：热门选择')}
                        showClear
                      />
                    </Col>

                    <Col span={12}>
                      <Form.Input
                        field='display_highlight_tag'
                        label={t('高亮标签')}
                        placeholder={t('例如：限时优惠')}
                        showClear
                      />
                    </Col>
                    <Col span={12}>
                      <Form.Input
                        field='display_cta_text'
                        label={t('按钮文案')}
                        placeholder={t('例如：立即开通')}
                        showClear
                      />
                    </Col>

                    <Col span={24}>
                      <Form.TextArea
                        field='display_subtitle'
                        label={t('展示副标题')}
                        placeholder={t('例如：适合高频使用与团队协作')}
                        autosize={{ minRows: 2, maxRows: 5 }}
                        showClear
                      />
                    </Col>

                    <Col span={12}>
                      <Form.Input
                        field='display_price_caption'
                        label={t('价格说明')}
                        placeholder={t('例如：按月自动续费')}
                        showClear
                      />
                    </Col>
                    <Col span={12}>
                      <Form.Input
                        field='display_quota_caption'
                        label={t('额度说明')}
                        placeholder={t('例如：每月专属额度')}
                        showClear
                      />
                    </Col>

                    <Col span={12}>
                      <Form.Select field='display_theme' label={t('展示主题')}>
                        {displayThemeOptions.map((option) => (
                          <Select.Option
                            key={option.value}
                            value={option.value}
                          >
                            {option.label}
                          </Select.Option>
                        ))}
                      </Form.Select>
                    </Col>
                    <Col span={12}>
                      <Form.Switch
                        field='display_featured'
                        label={t('推荐套餐')}
                        size='large'
                      />
                    </Col>

                    {displayFeatureFields.map((field, index) => (
                      <Col span={24} key={field}>
                        <Form.Input
                          field={field}
                          label={`${t('卖点')} ${index + 1}`}
                          placeholder={t('例如：低延迟、高可用、优先模型访问')}
                          showClear
                        />
                      </Col>
                    ))}

                    <Col span={24}>
                      <Form.TextArea
                        field='display_notes'
                        label={t('自由补充文案')}
                        placeholder={t(
                          '可填写最多 5 行的补充说明，用于前端套餐卡片展示',
                        )}
                        autosize={{ minRows: 3, maxRows: 6 }}
                        showClear
                      />
                    </Col>
                  </Row>
                </Card>

                {/* 有效期设置 */}
                <Card className='!rounded-2xl shadow-sm border-0 mb-4'>
                  <div className='flex items-center mb-2'>
                    <Avatar
                      size='small'
                      color='green'
                      className='mr-2 shadow-md'
                    >
                      <Clock size={16} />
                    </Avatar>
                    <div>
                      <Text className='text-lg font-medium'>
                        {t('有效期设置')}
                      </Text>
                      <div className='text-xs text-gray-600'>
                        {t('配置套餐的有效时长')}
                      </div>
                    </div>
                  </div>

                  <Row gutter={12}>
                    <Col span={12}>
                      <Form.Select
                        field='duration_unit'
                        label={t('有效期单位')}
                        required
                        rules={[{ required: true }]}
                      >
                        {durationUnitOptions.map((o) => (
                          <Select.Option key={o.value} value={o.value}>
                            {o.label}
                          </Select.Option>
                        ))}
                      </Form.Select>
                    </Col>

                    <Col span={12}>
                      {values.duration_unit === 'custom' ? (
                        <Form.InputNumber
                          field='custom_seconds'
                          label={t('自定义秒数')}
                          required
                          min={1}
                          precision={0}
                          rules={[{ required: true, message: t('请输入秒数') }]}
                          style={{ width: '100%' }}
                        />
                      ) : (
                        <Form.InputNumber
                          field='duration_value'
                          label={t('有效期数值')}
                          required
                          min={1}
                          precision={0}
                          rules={[{ required: true, message: t('请输入数值') }]}
                          style={{ width: '100%' }}
                        />
                      )}
                    </Col>
                  </Row>
                </Card>

                {/* 额度重置 */}
                <Card className='!rounded-2xl shadow-sm border-0 mb-4'>
                  <div className='flex items-center mb-2'>
                    <Avatar
                      size='small'
                      color='orange'
                      className='mr-2 shadow-md'
                    >
                      <RefreshCw size={16} />
                    </Avatar>
                    <div>
                      <Text className='text-lg font-medium'>
                        {t('额度重置')}
                      </Text>
                      <div className='text-xs text-gray-600'>
                        {t('支持周期性重置套餐权益额度')}
                      </div>
                    </div>
                  </div>

                  <Row gutter={12}>
                    <Col span={12}>
                      <Form.Select
                        field='quota_reset_period'
                        label={t('重置周期')}
                      >
                        {resetPeriodOptions.map((o) => (
                          <Select.Option key={o.value} value={o.value}>
                            {o.label}
                          </Select.Option>
                        ))}
                      </Form.Select>
                    </Col>
                    <Col span={12}>
                      {values.quota_reset_period === 'custom' ? (
                        <Form.InputNumber
                          field='quota_reset_custom_seconds'
                          label={t('自定义秒数')}
                          required
                          min={60}
                          precision={0}
                          rules={[{ required: true, message: t('请输入秒数') }]}
                          style={{ width: '100%' }}
                        />
                      ) : (
                        <Form.InputNumber
                          field='quota_reset_custom_seconds'
                          label={t('自定义秒数')}
                          min={0}
                          precision={0}
                          style={{ width: '100%' }}
                          disabled
                        />
                      )}
                    </Col>
                  </Row>
                </Card>

                {/* 第三方支付配置 */}
                <Card className='!rounded-2xl shadow-sm border-0 mb-4'>
                  <div className='flex items-center mb-2'>
                    <Avatar
                      size='small'
                      color='purple'
                      className='mr-2 shadow-md'
                    >
                      <IconCreditCard size={16} />
                    </Avatar>
                    <div>
                      <Text className='text-lg font-medium'>
                        {t('第三方支付配置')}
                      </Text>
                      <div className='text-xs text-gray-600'>
                        {t('Stripe/Creem 商品ID（可选）')}
                      </div>
                    </div>
                  </div>

                  <Row gutter={12}>
                    <Col span={24}>
                      <Form.Input
                        field='stripe_price_id'
                        label='Stripe PriceId'
                        placeholder='price_...'
                        showClear
                      />
                    </Col>

                    <Col span={24}>
                      <Form.Input
                        field='creem_product_id'
                        label='Creem ProductId'
                        placeholder='prod_...'
                        showClear
                      />
                    </Col>
                  </Row>
                </Card>
              </div>
            )}
          </Form>
        </Spin>
      </SideSheet>
    </>
  );
};

export default AddEditSubscriptionModal;
