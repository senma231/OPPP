import React, { useEffect, useState } from 'react';
import { Table, Card, Button, Tag, Space, Modal, Form, Input, Select, InputNumber, message } from 'antd';
import { PlusOutlined, ReloadOutlined, EditOutlined, DeleteOutlined, PlayCircleOutlined, PauseCircleOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import axios from 'axios';
import { API_URL } from '../config';

const { Option } = Select;

interface Forward {
  id: string;
  protocol: 'tcp' | 'udp';
  srcPort: number;
  dstHost: string;
  dstPort: number;
  description: string;
  enabled: boolean;
  stats: {
    bytesSent: number;
    bytesReceived: number;
    connections: number;
    startTime: string;
  };
}

const Forwards: React.FC = () => {
  const [forwards, setForwards] = useState<Forward[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalVisible, setModalVisible] = useState(false);
  const [form] = Form.useForm();
  const navigate = useNavigate();

  const fetchForwards = async () => {
    setLoading(true);
    try {
      const token = localStorage.getItem('token');
      const response = await axios.get(`${API_URL}/forwards`, {
        headers: { Authorization: `Bearer ${token}` }
      });
      setForwards(response.data.forwards || []);
    } catch (error: any) {
      message.error(error.response?.data?.error || '获取转发规则列表失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchForwards();
  }, []);

  const handleAddForward = async (values: any) => {
    try {
      const token = localStorage.getItem('token');
      await axios.post(`${API_URL}/forwards`, values, {
        headers: { Authorization: `Bearer ${token}` }
      });
      message.success('添加转发规则成功');
      setModalVisible(false);
      form.resetFields();
      fetchForwards();
    } catch (error: any) {
      message.error(error.response?.data?.error || '添加转发规则失败');
    }
  };

  const handleDeleteForward = async (id: string) => {
    Modal.confirm({
      title: '确认删除',
      content: '确定要删除这个转发规则吗？',
      onOk: async () => {
        try {
          const token = localStorage.getItem('token');
          await axios.delete(`${API_URL}/forwards/${id}`, {
            headers: { Authorization: `Bearer ${token}` }
          });
          message.success('删除转发规则成功');
          fetchForwards();
        } catch (error: any) {
          message.error(error.response?.data?.error || '删除转发规则失败');
        }
      }
    });
  };

  const handleEnableForward = async (id: string) => {
    try {
      const token = localStorage.getItem('token');
      await axios.post(`${API_URL}/forwards/${id}/enable`, {}, {
        headers: { Authorization: `Bearer ${token}` }
      });
      message.success('启用转发规则成功');
      fetchForwards();
    } catch (error: any) {
      message.error(error.response?.data?.error || '启用转发规则失败');
    }
  };

  const handleDisableForward = async (id: string) => {
    try {
      const token = localStorage.getItem('token');
      await axios.post(`${API_URL}/forwards/${id}/disable`, {}, {
        headers: { Authorization: `Bearer ${token}` }
      });
      message.success('禁用转发规则成功');
      fetchForwards();
    } catch (error: any) {
      message.error(error.response?.data?.error || '禁用转发规则失败');
    }
  };

  const formatBytes = (bytes: number) => {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  };

  const columns = [
    {
      title: '描述',
      dataIndex: 'description',
      key: 'description',
      render: (text: string, record: Forward) => (
        <a onClick={() => navigate(`/forwards/${record.id}`)}>{text}</a>
      )
    },
    {
      title: '协议',
      dataIndex: 'protocol',
      key: 'protocol',
      render: (text: string) => text.toUpperCase()
    },
    {
      title: '源端口',
      dataIndex: 'srcPort',
      key: 'srcPort',
    },
    {
      title: '目标主机',
      dataIndex: 'dstHost',
      key: 'dstHost',
    },
    {
      title: '目标端口',
      dataIndex: 'dstPort',
      key: 'dstPort',
    },
    {
      title: '状态',
      dataIndex: 'enabled',
      key: 'enabled',
      render: (enabled: boolean) => (
        <Tag color={enabled ? 'success' : 'default'}>
          {enabled ? '已启用' : '已禁用'}
        </Tag>
      )
    },
    {
      title: '流量',
      key: 'traffic',
      render: (_: any, record: Forward) => (
        <span>
          ↑ {formatBytes(record.stats.bytesSent)} / ↓ {formatBytes(record.stats.bytesReceived)}
        </span>
      )
    },
    {
      title: '连接数',
      dataIndex: ['stats', 'connections'],
      key: 'connections',
    },
    {
      title: '操作',
      key: 'action',
      render: (_: any, record: Forward) => (
        <Space size="middle">
          {record.enabled ? (
            <Button 
              type="text" 
              icon={<PauseCircleOutlined />} 
              onClick={() => handleDisableForward(record.id)}
            />
          ) : (
            <Button 
              type="text" 
              icon={<PlayCircleOutlined />} 
              onClick={() => handleEnableForward(record.id)}
            />
          )}
          <Button 
            type="text" 
            icon={<EditOutlined />} 
            onClick={() => navigate(`/forwards/${record.id}`)}
          />
          <Button 
            type="text" 
            danger 
            icon={<DeleteOutlined />} 
            onClick={() => handleDeleteForward(record.id)}
          />
        </Space>
      ),
    },
  ];

  return (
    <div>
      <Card
        title="端口转发管理"
        extra={
          <Space>
            <Button 
              type="primary" 
              icon={<PlusOutlined />} 
              onClick={() => setModalVisible(true)}
            >
              添加转发规则
            </Button>
            <Button 
              icon={<ReloadOutlined />} 
              onClick={fetchForwards}
              loading={loading}
            >
              刷新
            </Button>
          </Space>
        }
      >
        <Table 
          columns={columns} 
          dataSource={forwards} 
          rowKey="id" 
          loading={loading}
          pagination={{ pageSize: 10 }}
        />
      </Card>

      <Modal
        title="添加转发规则"
        visible={modalVisible}
        onCancel={() => setModalVisible(false)}
        footer={null}
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={handleAddForward}
        >
          <Form.Item
            name="description"
            label="描述"
            rules={[{ required: true, message: '请输入描述' }]}
          >
            <Input placeholder="请输入描述" />
          </Form.Item>
          <Form.Item
            name="protocol"
            label="协议"
            rules={[{ required: true, message: '请选择协议' }]}
          >
            <Select placeholder="请选择协议">
              <Option value="tcp">TCP</Option>
              <Option value="udp">UDP</Option>
            </Select>
          </Form.Item>
          <Form.Item
            name="srcPort"
            label="源端口"
            rules={[{ required: true, message: '请输入源端口' }]}
          >
            <InputNumber min={1} max={65535} style={{ width: '100%' }} placeholder="请输入源端口" />
          </Form.Item>
          <Form.Item
            name="dstHost"
            label="目标主机"
            rules={[{ required: true, message: '请输入目标主机' }]}
          >
            <Input placeholder="请输入目标主机，如 localhost 或 192.168.1.5" />
          </Form.Item>
          <Form.Item
            name="dstPort"
            label="目标端口"
            rules={[{ required: true, message: '请输入目标端口' }]}
          >
            <InputNumber min={1} max={65535} style={{ width: '100%' }} placeholder="请输入目标端口" />
          </Form.Item>
          <Form.Item>
            <Button type="primary" htmlType="submit" block>
              添加
            </Button>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default Forwards;
