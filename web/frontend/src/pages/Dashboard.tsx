import React, { useEffect, useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { Row, Col, Card, Statistic, Table, Tag, Button, Space, message } from 'antd';
import { DesktopOutlined, AppstoreOutlined, SwapOutlined, ReloadOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { RootState } from '../store';
import { fetchDevices } from '../store/slices/deviceSlice';
import { fetchApps } from '../store/slices/appSlice';
import { fetchForwards } from '../store/slices/forwardSlice';
import axios from 'axios';
import { API_URL } from '../config';
import ReactECharts from 'echarts-for-react';

interface SystemStatus {
  version: string;
  uptime: number;
  devices: {
    total: number;
    online: number;
  };
  apps: {
    total: number;
    running: number;
  };
  forwards: {
    total: number;
    enabled: number;
  };
  connections: {
    direct: number;
    upnp: number;
    holePunch: number;
    relay: number;
  };
  traffic: {
    sent: number;
    received: number;
  };
}

const Dashboard: React.FC = () => {
  const dispatch = useDispatch();
  const navigate = useNavigate();
  const { devices } = useSelector((state: RootState) => state.device);
  const { apps } = useSelector((state: RootState) => state.app);
  const { forwards } = useSelector((state: RootState) => state.forward);
  const [status, setStatus] = useState<SystemStatus | null>(null);
  const [loading, setLoading] = useState(false);

  const fetchData = async () => {
    setLoading(true);
    dispatch(fetchDevices());
    dispatch(fetchApps());
    dispatch(fetchForwards());

    try {
      const token = localStorage.getItem('token');
      const response = await axios.get(`${API_URL}/status`, {
        headers: { Authorization: `Bearer ${token}` }
      });
      setStatus(response.data);
    } catch (error: any) {
      message.error(error.response?.data?.error || '获取系统状态失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchData();
  }, [dispatch]);

  const onlineDevices = devices.filter(device => device.status === 'online').length;
  const runningApps = apps.filter(app => app.status === 'running').length;
  const enabledForwards = forwards.filter(forward => forward.enabled).length;

  const deviceColumns = [
    {
      title: '设备名称',
      dataIndex: 'name',
      key: 'name',
      render: (text: string, record: any) => (
        <a onClick={() => navigate(`/devices/${record.id}`)}>{text}</a>
      ),
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => (
        <Tag color={status === 'online' ? 'success' : 'error'}>
          {status === 'online' ? '在线' : '离线'}
        </Tag>
      ),
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
  ];

  const appColumns = [
    {
      title: '应用名称',
      dataIndex: 'name',
      key: 'name',
      render: (text: string, record: any) => (
        <a onClick={() => navigate(`/apps/${record.id}`)}>{text}</a>
      ),
    },
    {
      title: '协议',
      dataIndex: 'protocol',
      key: 'protocol',
      render: (protocol: string) => protocol.toUpperCase(),
    },
    {
      title: '本地端口',
      dataIndex: 'srcPort',
      key: 'srcPort',
    },
    {
      title: '目标设备',
      dataIndex: 'peerNode',
      key: 'peerNode',
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => (
        <Tag color={status === 'running' ? 'success' : status === 'stopped' ? 'warning' : 'error'}>
          {status === 'running' ? '运行中' : status === 'stopped' ? '已停止' : '错误'}
        </Tag>
      ),
    },
  ];

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

  const connectionOption = {
    tooltip: {
      trigger: 'item'
    },
    legend: {
      top: '5%',
      left: 'center'
    },
    series: [
      {
        name: '连接类型',
        type: 'pie',
        radius: ['40%', '70%'],
        avoidLabelOverlap: false,
        itemStyle: {
          borderRadius: 10,
          borderColor: '#fff',
          borderWidth: 2
        },
        label: {
          show: false,
          position: 'center'
        },
        emphasis: {
          label: {
            show: true,
            fontSize: '18',
            fontWeight: 'bold'
          }
        },
        labelLine: {
          show: false
        },
        data: status ? [
          { value: status.connections.direct, name: '直接连接' },
          { value: status.connections.upnp, name: 'UPnP' },
          { value: status.connections.holePunch, name: '打洞' },
          { value: status.connections.relay, name: '中继' }
        ] : []
      }
    ]
  };

  return (
    <div>
      <div style={{ marginBottom: 16 }}>
        <Button
          icon={<ReloadOutlined />}
          onClick={fetchData}
          loading={loading}
        >
          刷新
        </Button>
      </div>

      <Row gutter={16}>
        <Col span={8}>
          <Card>
            <Statistic
              title="设备总数"
              value={devices.length}
              prefix={<DesktopOutlined />}
              suffix={<span style={{ fontSize: 14 }}>{`/${onlineDevices} 在线`}</span>}
            />
          </Card>
        </Col>
        <Col span={8}>
          <Card>
            <Statistic
              title="应用总数"
              value={apps.length}
              prefix={<AppstoreOutlined />}
              suffix={<span style={{ fontSize: 14 }}>{`/${runningApps} 运行中`}</span>}
            />
          </Card>
        </Col>
        <Col span={8}>
          <Card>
            <Statistic
              title="转发规则总数"
              value={forwards.length}
              prefix={<SwapOutlined />}
              suffix={<span style={{ fontSize: 14 }}>{`/${enabledForwards} 已启用`}</span>}
            />
          </Card>
        </Col>
      </Row>

      <Row gutter={16} style={{ marginTop: 16 }}>
        <Col span={12}>
          <Card title="系统信息">
            <p><strong>版本：</strong> {status?.version || '未知'}</p>
            <p><strong>运行时间：</strong> {status ? formatTime(status.uptime) : '未知'}</p>
            <p><strong>总发送流量：</strong> {status ? formatBytes(status.traffic.sent) : '0 B'}</p>
            <p><strong>总接收流量：</strong> {status ? formatBytes(status.traffic.received) : '0 B'}</p>
          </Card>
        </Col>
        <Col span={12}>
          <Card title="连接类型分布">
            <ReactECharts option={connectionOption} style={{ height: 300 }} />
          </Card>
        </Col>
      </Row>

      <Card title="设备状态" style={{ marginTop: 16 }}>
        <Table
          dataSource={devices}
          columns={deviceColumns}
          rowKey="id"
          pagination={false}
          loading={loading}
        />
      </Card>

      <Card title="应用状态" style={{ marginTop: 16 }}>
        <Table
          dataSource={apps}
          columns={appColumns}
          rowKey="id"
          pagination={false}
          loading={loading}
        />
      </Card>
    </div>
  );
};

export default Dashboard;
