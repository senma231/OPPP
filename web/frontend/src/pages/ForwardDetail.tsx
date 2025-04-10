import React, { useEffect, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Card, Descriptions, Button, Tabs, Statistic, Row, Col, Tag, Space, message } from 'antd';
import { ArrowLeftOutlined, ReloadOutlined, EditOutlined, PlayCircleOutlined, PauseCircleOutlined } from '@ant-design/icons';
import axios from 'axios';
import { API_URL } from '../config';
import ReactECharts from 'echarts-for-react';

const { TabPane } = Tabs;

interface Forward {
  id: string;
  protocol: 'tcp' | 'udp';
  srcPort: number;
  dstHost: string;
  dstPort: number;
  description: string;
  enabled: boolean;
  createdAt: string;
  updatedAt: string;
  stats: {
    bytesSent: number;
    bytesReceived: number;
    connections: number;
    startTime: string;
  };
}

const ForwardDetail: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [forward, setForward] = useState<Forward | null>(null);
  const [loading, setLoading] = useState(false);

  const fetchForward = async () => {
    if (!id) return;
    
    setLoading(true);
    try {
      const token = localStorage.getItem('token');
      const response = await axios.get(`${API_URL}/forwards/${id}`, {
        headers: { Authorization: `Bearer ${token}` }
      });
      setForward(response.data);
    } catch (error: any) {
      message.error(error.response?.data?.error || '获取转发规则详情失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchForward();
  }, [id]);

  const handleEnableForward = async () => {
    if (!id) return;
    
    try {
      const token = localStorage.getItem('token');
      await axios.post(`${API_URL}/forwards/${id}/enable`, {}, {
        headers: { Authorization: `Bearer ${token}` }
      });
      message.success('启用转发规则成功');
      fetchForward();
    } catch (error: any) {
      message.error(error.response?.data?.error || '启用转发规则失败');
    }
  };

  const handleDisableForward = async () => {
    if (!id) return;
    
    try {
      const token = localStorage.getItem('token');
      await axios.post(`${API_URL}/forwards/${id}/disable`, {}, {
        headers: { Authorization: `Bearer ${token}` }
      });
      message.success('禁用转发规则成功');
      fetchForward();
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

  const getTrafficOption = () => {
    return {
      tooltip: {
        trigger: 'item',
        formatter: '{a} <br/>{b} : {c} ({d}%)'
      },
      legend: {
        orient: 'vertical',
        left: 'left',
        data: ['发送流量', '接收流量']
      },
      series: [
        {
          name: '流量统计',
          type: 'pie',
          radius: '55%',
          center: ['50%', '60%'],
          data: [
            { value: forward?.stats.bytesSent || 0, name: '发送流量' },
            { value: forward?.stats.bytesReceived || 0, name: '接收流量' }
          ],
          emphasis: {
            itemStyle: {
              shadowBlur: 10,
              shadowOffsetX: 0,
              shadowColor: 'rgba(0, 0, 0, 0.5)'
            }
          }
        }
      ]
    };
  };

  if (!forward) {
    return (
      <Card loading={loading} title="转发规则详情">
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
              onClick={() => navigate('/forwards')}
              type="text"
            />
            转发规则详情
          </Space>
        }
        extra={
          <Space>
            <Button 
              icon={<ReloadOutlined />} 
              onClick={fetchForward}
              loading={loading}
            >
              刷新
            </Button>
            {forward.enabled ? (
              <Button 
                icon={<PauseCircleOutlined />} 
                onClick={handleDisableForward}
              >
                禁用
              </Button>
            ) : (
              <Button 
                type="primary" 
                icon={<PlayCircleOutlined />} 
                onClick={handleEnableForward}
              >
                启用
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
          <Descriptions.Item label="描述">{forward.description}</Descriptions.Item>
          <Descriptions.Item label="状态">
            <Tag color={forward.enabled ? 'success' : 'default'}>
              {forward.enabled ? '已启用' : '已禁用'}
            </Tag>
          </Descriptions.Item>
          <Descriptions.Item label="协议">{forward.protocol.toUpperCase()}</Descriptions.Item>
          <Descriptions.Item label="源端口">{forward.srcPort}</Descriptions.Item>
          <Descriptions.Item label="目标主机">{forward.dstHost}</Descriptions.Item>
          <Descriptions.Item label="目标端口">{forward.dstPort}</Descriptions.Item>
          <Descriptions.Item label="创建时间">{new Date(forward.createdAt).toLocaleString()}</Descriptions.Item>
          <Descriptions.Item label="更新时间">{new Date(forward.updatedAt).toLocaleString()}</Descriptions.Item>
        </Descriptions>

        <Tabs defaultActiveKey="stats" style={{ marginTop: 16 }}>
          <TabPane tab="统计信息" key="stats">
            <Row gutter={16}>
              <Col span={12}>
                <Card>
                  <Row gutter={16}>
                    <Col span={12}>
                      <Statistic
                        title="发送流量"
                        value={formatBytes(forward.stats.bytesSent)}
                        valueStyle={{ color: '#3f8600' }}
                      />
                    </Col>
                    <Col span={12}>
                      <Statistic
                        title="接收流量"
                        value={formatBytes(forward.stats.bytesReceived)}
                        valueStyle={{ color: '#3f8600' }}
                      />
                    </Col>
                  </Row>
                  <Row gutter={16} style={{ marginTop: 16 }}>
                    <Col span={12}>
                      <Statistic
                        title="连接数"
                        value={forward.stats.connections}
                      />
                    </Col>
                    <Col span={12}>
                      <Statistic
                        title="开始时间"
                        value={new Date(forward.stats.startTime).toLocaleString()}
                      />
                    </Col>
                  </Row>
                </Card>
              </Col>
              <Col span={12}>
                <Card title="流量分布">
                  <ReactECharts option={getTrafficOption()} style={{ height: 300 }} />
                </Card>
              </Col>
            </Row>
          </TabPane>
          <TabPane tab="使用说明" key="usage">
            <Card>
              <h3>如何使用此转发规则</h3>
              <p>此转发规则将本地端口 {forward.srcPort} 转发到 {forward.dstHost}:{forward.dstPort}。</p>
              <p>您可以通过以下方式访问远程服务：</p>
              <ul>
                <li>在本地访问 <strong>localhost:{forward.srcPort}</strong> 或 <strong>127.0.0.1:{forward.srcPort}</strong></li>
              </ul>
              <p>例如，如果这是一个远程桌面转发，您可以在远程桌面客户端中输入 <strong>localhost:{forward.srcPort}</strong> 进行连接。</p>
            </Card>
          </TabPane>
        </Tabs>
      </Card>
    </div>
  );
};

export default ForwardDetail;
