import React, { useEffect, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Card, Descriptions, Button, Tabs, Table, Tag, Space, Statistic, Row, Col, message } from 'antd';
import { ArrowLeftOutlined, ReloadOutlined, EditOutlined, PlusOutlined } from '@ant-design/icons';
import axios from 'axios';
import { API_URL } from '../config';

const { TabPane } = Tabs;

interface Device {
  id: string;
  name: string;
  nodeId: string;
  status: 'online' | 'offline';
  natType: string;
  externalIP: string;
  localIP: string;
  version: string;
  os: string;
  arch: string;
  lastSeenAt: string;
}

interface App {
  id: string;
  name: string;
  protocol: string;
  srcPort: number;
  peerNode: string;
  dstPort: number;
  dstHost: string;
  status: string;
  description: string;
}

interface Stats {
  bytesSent: number;
  bytesReceived: number;
  connections: number;
  connectionTime: number;
}

const DeviceDetail: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [device, setDevice] = useState<Device | null>(null);
  const [apps, setApps] = useState<App[]>([]);
  const [stats, setStats] = useState<Stats | null>(null);
  const [loading, setLoading] = useState(false);
  const [appsLoading, setAppsLoading] = useState(false);

  const fetchDevice = async () => {
    if (!id) return;
    
    setLoading(true);
    try {
      const token = localStorage.getItem('token');
      const response = await axios.get(`${API_URL}/devices/${id}`, {
        headers: { Authorization: `Bearer ${token}` }
      });
      setDevice(response.data);
    } catch (error: any) {
      message.error(error.response?.data?.error || '获取设备详情失败');
    } finally {
      setLoading(false);
    }
  };

  const fetchApps = async () => {
    if (!id) return;
    
    setAppsLoading(true);
    try {
      const token = localStorage.getItem('token');
      const response = await axios.get(`${API_URL}/apps?deviceId=${id}`, {
        headers: { Authorization: `Bearer ${token}` }
      });
      setApps(response.data.apps || []);
    } catch (error: any) {
      message.error(error.response?.data?.error || '获取应用列表失败');
    } finally {
      setAppsLoading(false);
    }
  };

  const fetchStats = async () => {
    if (!id) return;
    
    try {
      const token = localStorage.getItem('token');
      const response = await axios.get(`${API_URL}/devices/${id}/stats`, {
        headers: { Authorization: `Bearer ${token}` }
      });
      setStats(response.data);
    } catch (error: any) {
      message.error(error.response?.data?.error || '获取统计信息失败');
    }
  };

  useEffect(() => {
    fetchDevice();
    fetchApps();
    fetchStats();
  }, [id]);

  const formatBytes = (bytes: number) => {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  };

  const formatTime = (seconds: number) => {
    const days = Math.floor(seconds / 86400);
    const hours = Math.floor((seconds % 86400) / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);
    
    if (days > 0) {
      return `${days}天 ${hours}小时`;
    } else if (hours > 0) {
      return `${hours}小时 ${minutes}分钟`;
    } else {
      return `${minutes}分钟`;
    }
  };

  const appColumns = [
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
          <Button 
            type="text" 
            icon={<EditOutlined />} 
            onClick={() => navigate(`/apps/${record.id}`)}
          />
        </Space>
      ),
    },
  ];

  if (!device) {
    return (
      <Card loading={loading} title="设备详情">
        <div>加载中...</div>
      </Card>
    );
  }

  return (
    <div>
      <Card
        title={
          <Space>
            <Button 
              icon={<ArrowLeftOutlined />} 
              onClick={() => navigate('/devices')}
              type="text"
            />
            设备详情
          </Space>
        }
        extra={
          <Space>
            <Button 
              icon={<ReloadOutlined />} 
              onClick={() => {
                fetchDevice();
                fetchApps();
                fetchStats();
              }}
              loading={loading}
            >
              刷新
            </Button>
            <Button 
              type="primary" 
              icon={<EditOutlined />} 
              onClick={() => message.info('编辑功能尚未实现')}
            >
              编辑
            </Button>
          </Space>
        }
        loading={loading}
      >
        <Descriptions bordered column={2}>
          <Descriptions.Item label="设备名称">{device.name}</Descriptions.Item>
          <Descriptions.Item label="节点 ID">{device.nodeId}</Descriptions.Item>
          <Descriptions.Item label="状态">
            <Tag color={device.status === 'online' ? 'success' : 'error'}>
              {device.status === 'online' ? '在线' : '离线'}
            </Tag>
          </Descriptions.Item>
          <Descriptions.Item label="NAT 类型">{device.natType}</Descriptions.Item>
          <Descriptions.Item label="外部 IP">{device.externalIP}</Descriptions.Item>
          <Descriptions.Item label="本地 IP">{device.localIP}</Descriptions.Item>
          <Descriptions.Item label="版本">{device.version}</Descriptions.Item>
          <Descriptions.Item label="操作系统">{device.os}</Descriptions.Item>
          <Descriptions.Item label="架构">{device.arch}</Descriptions.Item>
          <Descriptions.Item label="最后在线时间">{new Date(device.lastSeenAt).toLocaleString()}</Descriptions.Item>
        </Descriptions>

        <Tabs defaultActiveKey="apps" style={{ marginTop: 16 }}>
          <TabPane tab="应用" key="apps">
            <div style={{ marginBottom: 16 }}>
              <Button 
                type="primary" 
                icon={<PlusOutlined />} 
                onClick={() => navigate('/apps/new', { state: { deviceId: id } })}
              >
                添加应用
              </Button>
            </div>
            <Table 
              columns={appColumns} 
              dataSource={apps} 
              rowKey="id" 
              loading={appsLoading}
              pagination={{ pageSize: 5 }}
            />
          </TabPane>
          <TabPane tab="统计信息" key="stats">
            {stats ? (
              <Row gutter={16}>
                <Col span={6}>
                  <Card>
                    <Statistic
                      title="发送流量"
                      value={formatBytes(stats.bytesSent)}
                      valueStyle={{ color: '#3f8600' }}
                    />
                  </Card>
                </Col>
                <Col span={6}>
                  <Card>
                    <Statistic
                      title="接收流量"
                      value={formatBytes(stats.bytesReceived)}
                      valueStyle={{ color: '#3f8600' }}
                    />
                  </Card>
                </Col>
                <Col span={6}>
                  <Card>
                    <Statistic
                      title="连接数"
                      value={stats.connections}
                    />
                  </Card>
                </Col>
                <Col span={6}>
                  <Card>
                    <Statistic
                      title="连接时长"
                      value={formatTime(stats.connectionTime)}
                    />
                  </Card>
                </Col>
              </Row>
            ) : (
              <div>暂无统计信息</div>
            )}
          </TabPane>
        </Tabs>
      </Card>
    </div>
  );
};

export default DeviceDetail;
