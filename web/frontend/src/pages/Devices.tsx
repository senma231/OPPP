import React, { useEffect, useState } from 'react';
import { Table, Card, Button, Tag, Space, Modal, Form, Input, message } from 'antd';
import { PlusOutlined, ReloadOutlined, EditOutlined, DeleteOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import axios from 'axios';
import { API_URL } from '../config';

interface Device {
  id: string;
  name: string;
  nodeId: string;
  status: 'online' | 'offline';
  natType: string;
  externalIP: string;
  lastSeenAt: string;
}

const Devices: React.FC = () => {
  const [devices, setDevices] = useState<Device[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalVisible, setModalVisible] = useState(false);
  const [form] = Form.useForm();
  const navigate = useNavigate();

  const fetchDevices = async () => {
    setLoading(true);
    try {
      const token = localStorage.getItem('token');
      const response = await axios.get(`${API_URL}/devices`, {
        headers: { Authorization: `Bearer ${token}` }
      });
      setDevices(response.data.devices || []);
    } catch (error: any) {
      message.error(error.response?.data?.error || '获取设备列表失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchDevices();
  }, []);

  const handleAddDevice = async (values: any) => {
    try {
      const token = localStorage.getItem('token');
      await axios.post(`${API_URL}/devices`, values, {
        headers: { Authorization: `Bearer ${token}` }
      });
      message.success('添加设备成功');
      setModalVisible(false);
      form.resetFields();
      fetchDevices();
    } catch (error: any) {
      message.error(error.response?.data?.error || '添加设备失败');
    }
  };

  const handleDeleteDevice = async (id: string) => {
    Modal.confirm({
      title: '确认删除',
      content: '确定要删除这个设备吗？这将同时删除与该设备相关的所有应用。',
      onOk: async () => {
        try {
          const token = localStorage.getItem('token');
          await axios.delete(`${API_URL}/devices/${id}`, {
            headers: { Authorization: `Bearer ${token}` }
          });
          message.success('删除设备成功');
          fetchDevices();
        } catch (error: any) {
          message.error(error.response?.data?.error || '删除设备失败');
        }
      }
    });
  };

  const columns = [
    {
      title: '设备名称',
      dataIndex: 'name',
      key: 'name',
      render: (text: string, record: Device) => (
        <a onClick={() => navigate(`/devices/${record.id}`)}>{text}</a>
      )
    },
    {
      title: '节点 ID',
      dataIndex: 'nodeId',
      key: 'nodeId',
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => (
        <Tag color={status === 'online' ? 'success' : 'error'}>
          {status === 'online' ? '在线' : '离线'}
        </Tag>
      )
    },
    {
      title: 'NAT 类型',
      dataIndex: 'natType',
      key: 'natType',
    },
    {
      title: '外部 IP',
      dataIndex: 'externalIP',
      key: 'externalIP',
    },
    {
      title: '最后在线时间',
      dataIndex: 'lastSeenAt',
      key: 'lastSeenAt',
      render: (text: string) => new Date(text).toLocaleString()
    },
    {
      title: '操作',
      key: 'action',
      render: (_: any, record: Device) => (
        <Space size="middle">
          <Button 
            type="text" 
            icon={<EditOutlined />} 
            onClick={() => navigate(`/devices/${record.id}`)}
          />
          <Button 
            type="text" 
            danger 
            icon={<DeleteOutlined />} 
            onClick={() => handleDeleteDevice(record.id)}
          />
        </Space>
      ),
    },
  ];

  return (
    <div>
      <Card
        title="设备管理"
        extra={
          <Space>
            <Button 
              type="primary" 
              icon={<PlusOutlined />} 
              onClick={() => setModalVisible(true)}
            >
              添加设备
            </Button>
            <Button 
              icon={<ReloadOutlined />} 
              onClick={fetchDevices}
              loading={loading}
            >
              刷新
            </Button>
          </Space>
        }
      >
        <Table 
          columns={columns} 
          dataSource={devices} 
          rowKey="id" 
          loading={loading}
          pagination={{ pageSize: 10 }}
        />
      </Card>

      <Modal
        title="添加设备"
        visible={modalVisible}
        onCancel={() => setModalVisible(false)}
        footer={null}
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={handleAddDevice}
        >
          <Form.Item
            name="name"
            label="设备名称"
            rules={[{ required: true, message: '请输入设备名称' }]}
          >
            <Input placeholder="请输入设备名称" />
          </Form.Item>
          <Form.Item
            name="nodeId"
            label="节点 ID"
            rules={[{ required: true, message: '请输入节点 ID' }]}
          >
            <Input placeholder="请输入节点 ID" />
          </Form.Item>
          <Form.Item
            name="token"
            label="认证令牌"
            rules={[{ required: true, message: '请输入认证令牌' }]}
          >
            <Input.Password placeholder="请输入认证令牌" />
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

export default Devices;
