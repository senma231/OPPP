import React, { useEffect, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Card, Descriptions, Button, Tabs, Statistic, Row, Col, Tag, Space, message } from 'antd';
import { ArrowLeftOutlined, ReloadOutlined, EditOutlined, PlayCircleOutlined, PauseCircleOutlined } from '@ant-design/icons';
import axios from 'axios';
import { API_URL } from '../config';

const { TabPane } = Tabs;

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
  deviceId: string;
}

interface Stats {
  bytesSent: number;
  bytesReceived: number;
  connections: number;
  connectionTime: number;
}

const AppDetail: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [app, setApp] = useState<App | null>(null);
  const [stats, setStats] = useState<Stats | null>(null);
  const [loading, setLoading] = useState(false);

  const fetchApp = async () => {
    if (!id) return;
    
    setLoading(true);
    try {
      const token = localStorage.getItem('token');
      const response = await axios.get(`${API_URL}/apps/${id}`, {
        headers: { Authorization: `Bearer ${token}` }
      });
      setApp(response.data);
    } catch (error: any) {
      message.error(error.response?.data?.error || '获取应用详情失败');
    } finally {
      setLoading(false);
    }
  };

  const fetchStats = async () => {
    if (!id) return;
    
    try {
      const token = localStorage.getItem('token');
      const response = await axios.get(`${API_URL}/apps/${id}/stats`, {
        headers: { Authorization: `Bearer ${token}` }
      });
      setStats(response.data);
    } catch (error: any) {
      message.error(error.response?.data?.error || '获取统计信息失败');
    }
  };

  useEffect(() => {
    fetchApp();
    fetchStats();
  }, [id]);

  const handleStartApp = async () => {
    if (!id) return;
    
    try {
      const token = localStorage.getItem('token');
      await axios.post(`${API_URL}/apps/${id}/start`, {}, {
        headers: { Authorization: `Bearer ${token}` }
      });
      message.success('启动应用成功');
      fetchApp();
    } catch (error: any) {
      message.error(error.response?.data?.error || '启动应用失败');
    }
  };

  const handleStopApp = async () => {
    if (!id) return;
    
    try {
      const token = localStorage.getItem('token');
      await axios.post(`${API_URL}/apps/${id}/stop`, {}, {
        headers: { Authorization: `Bearer ${token}` }
      });
      message.success('停止应用成功');
      fetchApp();
    } catch (error: any) {
      message.error(error.response?.data?.error || '停止应用失败');
    }
  };

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

  if (!app) {
    return (
      <Card loading={loading} title="应用详情">
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
              onClick={() => navigate('/apps')}
              type="text"
            />
            应用详情
          </Space>
        }
        extra={
          <Space>
            <Button 
              icon={<ReloadOutlined />} 
              onClick={() => {
                fetchApp();
                fetchStats();
              }}
              loading={loading}
            >
              刷新
            </Button>
            {app.status === 'running' ? (
              <Button 
                icon={<PauseCircleOutlined />} 
                onClick={handleStopApp}
              >
                停止
              </Button>
            ) : (
              <Button 
                type="primary" 
                icon={<PlayCircleOutlined />} 
                onClick={handleStartApp}
              >
                启动
              </Button>
            )}
            <Button 
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
          <Descriptions.Item label="应用名称">{app.name}</Descriptions.Item>
          <Descriptions.Item label="状态">
            <Tag color={app.status === 'running' ? 'success' : app.status === 'stopped' ? 'warning' : 'error'}>
              {app.status === 'running' ? '运行中' : app.status === 'stopped' ? '已停止' : '错误'}
            </Tag>
          </Descriptions.Item>
          <Descriptions.Item label="协议">{app.protocol.toUpperCase()}</Descriptions.Item>
          <Descriptions.Item label="本地端口">{app.srcPort}</Descriptions.Item>
          <Descriptions.Item label="目标节点">{app.peerNode}</Descriptions.Item>
          <Descriptions.Item label="目标端口">{app.dstPort}</Descriptions.Item>
          <Descriptions.Item label="目标主机">{app.dstHost}</Descriptions.Item>
          <Descriptions.Item label="描述">{app.description || '无'}</Descriptions.Item>
        </Descriptions>

        <Tabs defaultActiveKey="stats" style={{ marginTop: 16 }}>
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
          <TabPane tab="使用说明" key="usage">
            <Card>
              <h3>如何使用此应用</h3>
              <p>此应用将本地端口 {app.srcPort} 转发到 {app.peerNode} 节点的 {app.dstHost}:{app.dstPort}。</p>
              <p>您可以通过以下方式访问远程服务：</p>
              <ul>
                <li>在本地访问 <strong>localhost:{app.srcPort}</strong> 或 <strong>127.0.0.1:{app.srcPort}</strong></li>
              </ul>
              <p>例如，如果这是一个远程桌面应用，您可以在远程桌面客户端中输入 <strong>localhost:{app.srcPort}</strong> 进行连接。</p>
            </Card>
          </TabPane>
        </Tabs>
      </Card>
    </div>
  );
};

export default AppDetail;
