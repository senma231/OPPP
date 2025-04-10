import React, { useEffect, useState } from 'react';
import { Table, Card, Button, Tag, Space, Modal, Form, Input, Select, InputNumber, message } from 'antd';
import { PlusOutlined, ReloadOutlined, EditOutlined, DeleteOutlined, PlayCircleOutlined, PauseCircleOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import axios from 'axios';
import { API_URL } from '../config';

const { Option } = Select;

interface Device {
  id: string;
  name: string;
}

interface App {
  id: string;
  name: string;
  protocol: 'tcp' | 'udp';
  srcPort: number;
  peerNode: string;
  dstPort: number;
  dstHost: string;
  status: 'running' | 'stopped' | 'error';
  description: string;
}

const Apps: React.FC = () => {
  const [apps, setApps] = useState<App[]>([]);
  const [devices, setDevices] = useState<Device[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalVisible, setModalVisible] = useState(false);
  const [form] = Form.useForm();
  const navigate = useNavigate();

  const fetchApps = async () => {
    setLoading(true);
    try {
      const token = localStorage.getItem('token');
      const response = await axios.get(`${API_URL}/apps`, {
        headers: { Authorization: `Bearer ${token}` }
      });
      setApps(response.data.apps || []);
    } catch (error: any) {
      message.error(error.response?.data?.error || '获取应用列表失败');
    } finally {
      setLoading(false);
    }
  };

  const fetchDevices = async () => {
    try {
      const token = localStorage.getItem('token');
      const response = await axios.get(`${API_URL}/devices`, {
        headers: { Authorization: `Bearer ${token}` }
      });
      setDevices(response.data.devices || []);
    } catch (error: any) {
      message.error(error.response?.data?.error || '获取设备列表失败');
    }
  };

  useEffect(() => {
    fetchApps();
    fetchDevices();
  }, []);

  const handleAddApp = async (values: any) => {
    try {
      const token = localStorage.getItem('token');
      await axios.post(`${API_URL}/apps`, values, {
        headers: { Authorization: `Bearer ${token}` }
      });
      message.success('添加应用成功');
      setModalVisible(false);
      form.resetFields();
      fetchApps();
    } catch (error: any) {
      message.error(error.response?.data?.error || '添加应用失败');
    }
  };

  const handleDeleteApp = async (id: string) => {
    Modal.confirm({
      title: '确认删除',
      content: '确定要删除这个应用吗？',
      onOk: async () => {
        try {
          const token = localStorage.getItem('token');
          await axios.delete(`${API_URL}/apps/${id}`, {
            headers: { Authorization: `Bearer ${token}` }
          });
          message.success('删除应用成功');
          fetchApps();
        } catch (error: any) {
          message.error(error.response?.data?.error || '删除应用失败');
        }
      }
    });
  };

  const handleStartApp = async (id: string) => {
    try {
      const token = localStorage.getItem('token');
      await axios.post(`${API_URL}/apps/${id}/start`, {}, {
        headers: { Authorization: `Bearer ${token}` }
      });
      message.success('启动应用成功');
      fetchApps();
    } catch (error: any) {
      message.error(error.response?.data?.error || '启动应用失败');
    }
  };

  const handleStopApp = async (id: string) => {
    try {
      const token = localStorage.getItem('token');
      await axios.post(`${API_URL}/apps/${id}/stop`, {}, {
        headers: { Authorization: `Bearer ${token}` }
      });
      message.success('停止应用成功');
      fetchApps();
    } catch (error: any) {
      message.error(error.response?.data?.error || '停止应用失败');
    }
  };

  const columns = [
    {
      title: '应用名称',
      dataIndex: 'name',
      key: 'name',
      render: (text: string, record: App) => (
        <a onClick={() => navigate(`/apps/${record.id}`)}>{text}</a>
      )
    },
    {
      title: '协议',
      dataIndex: 'protocol',
      key: 'protocol',
      render: (text: string) => text.toUpperCase()
    },
    {
      title: '本地端口',
      dataIndex: 'srcPort',
      key: 'srcPort',
    },
    {
      title: '目标节点',
      dataIndex: 'peerNode',
      key: 'peerNode',
    },
    {
      title: '目标端口',
      dataIndex: 'dstPort',
      key: 'dstPort',
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => (
        <Tag color={status === 'running' ? 'success' : status === 'stopped' ? 'warning' : 'error'}>
          {status === 'running' ? '运行中' : status === 'stopped' ? '已停止' : '错误'}
        </Tag>
      )
    },
    {
      title: '操作',
      key: 'action',
      render: (_: any, record: App) => (
        <Space size="middle">
          {record.status === 'running' ? (
            <Button 
              type="text" 
              icon={<PauseCircleOutlined />} 
              onClick={() => handleStopApp(record.id)}
            />
          ) : (
            <Button 
              type="text" 
              icon={<PlayCircleOutlined />} 
              onClick={() => handleStartApp(record.id)}
            />
          )}
          <Button 
            type="text" 
            icon={<EditOutlined />} 
            onClick={() => navigate(`/apps/${record.id}`)}
          />
          <Button 
            type="text" 
            danger 
            icon={<DeleteOutlined />} 
            onClick={() => handleDeleteApp(record.id)}
          />
        </Space>
      ),
    },
  ];

  return (
    <div>
      <Card
        title="应用管理"
        extra={
          <Space>
            <Button 
              type="primary" 
              icon={<PlusOutlined />} 
              onClick={() => setModalVisible(true)}
            >
              添加应用
            </Button>
            <Button 
              icon={<ReloadOutlined />} 
              onClick={fetchApps}
              loading={loading}
            >
              刷新
            </Button>
          </Space>
        }
      >
        <Table 
          columns={columns} 
          dataSource={apps} 
          rowKey="id" 
          loading={loading}
          pagination={{ pageSize: 10 }}
        />
      </Card>

      <Modal
        title="添加应用"
        visible={modalVisible}
        onCancel={() => setModalVisible(false)}
        footer={null}
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={handleAddApp}
        >
          <Form.Item
            name="deviceId"
            label="设备"
            rules={[{ required: true, message: '请选择设备' }]}
          >
            <Select placeholder="请选择设备">
              {devices.map(device => (
                <Option key={device.id} value={device.id}>{device.name}</Option>
              ))}
            </Select>
          </Form.Item>
          <Form.Item
            name="name"
            label="应用名称"
            rules={[{ required: true, message: '请输入应用名称' }]}
          >
            <Input placeholder="请输入应用名称" />
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
            label="本地端口"
            rules={[{ required: true, message: '请输入本地端口' }]}
          >
            <InputNumber min={1} max={65535} style={{ width: '100%' }} placeholder="请输入本地端口" />
          </Form.Item>
          <Form.Item
            name="peerNode"
            label="目标节点"
            rules={[{ required: true, message: '请输入目标节点' }]}
          >
            <Input placeholder="请输入目标节点" />
          </Form.Item>
          <Form.Item
            name="dstPort"
            label="目标端口"
            rules={[{ required: true, message: '请输入目标端口' }]}
          >
            <InputNumber min={1} max={65535} style={{ width: '100%' }} placeholder="请输入目标端口" />
          </Form.Item>
          <Form.Item
            name="dstHost"
            label="目标主机"
            rules={[{ required: true, message: '请输入目标主机' }]}
          >
            <Input placeholder="请输入目标主机，如 localhost 或 192.168.1.5" />
          </Form.Item>
          <Form.Item
            name="description"
            label="描述"
          >
            <Input.TextArea placeholder="请输入应用描述" />
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

export default Apps;
