import React, { useEffect } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { Row, Col, Card, Statistic, Table, Tag } from 'antd';
import { DesktopOutlined, AppstoreOutlined, SwapOutlined, CheckCircleOutlined, CloseCircleOutlined } from '@ant-design/icons';
import { RootState } from '../store';
import { fetchDevices } from '../store/slices/deviceSlice';
import { fetchApps } from '../store/slices/appSlice';
import { fetchForwards } from '../store/slices/forwardSlice';

const Dashboard: React.FC = () => {
  const dispatch = useDispatch();
  const { devices } = useSelector((state: RootState) => state.device);
  const { apps } = useSelector((state: RootState) => state.app);
  const { forwards } = useSelector((state: RootState) => state.forward);

  useEffect(() => {
    dispatch(fetchDevices());
    dispatch(fetchApps());
    dispatch(fetchForwards());
  }, [dispatch]);

  const onlineDevices = devices.filter(device => device.status === 'online').length;
  const runningApps = apps.filter(app => app.status === 'running').length;
  const enabledForwards = forwards.filter(forward => forward.enabled).length;

  const deviceColumns = [
    {
      title: '设备名称',
      dataIndex: 'name',
      key: 'name',
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

  return (
    <div>
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

      <Card title="设备状态" style={{ marginTop: 16 }}>
        <Table
          dataSource={devices}
          columns={deviceColumns}
          rowKey="id"
          pagination={false}
        />
      </Card>

      <Card title="应用状态" style={{ marginTop: 16 }}>
        <Table
          dataSource={apps}
          columns={appColumns}
          rowKey="id"
          pagination={false}
        />
      </Card>
    </div>
  );
};

export default Dashboard;
