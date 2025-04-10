import React, { useState, useEffect } from 'react';
import { Card, Tabs, Form, Input, Button, Switch, InputNumber, message } from 'antd';
import { useSelector } from 'react-redux';
import { RootState } from '../store';
import axios from 'axios';
import { API_URL } from '../config';

const { TabPane } = Tabs;

interface UserSettings {
  email: string;
  notifications: boolean;
}

interface SystemSettings {
  serverAddress: string;
  serverPort: number;
  enableUPnP: boolean;
  enableRelay: boolean;
  maxConnections: number;
}

const Settings: React.FC = () => {
  const [userForm] = Form.useForm();
  const [systemForm] = Form.useForm();
  const [loading, setLoading] = useState(false);
  const { user } = useSelector((state: RootState) => state.auth);

  useEffect(() => {
    // 初始化用户设置表单
    if (user) {
      userForm.setFieldsValue({
        email: user.email,
        notifications: true,
      });
    }

    // 初始化系统设置表单
    systemForm.setFieldsValue({
      serverAddress: 'stun.example.com',
      serverPort: 3478,
      enableUPnP: true,
      enableRelay: true,
      maxConnections: 100,
    });
  }, [user, userForm, systemForm]);

  const handleUserSettingsSave = (values: UserSettings) => {
    setLoading(true);
    // 模拟 API 调用
    setTimeout(() => {
      message.success('用户设置已保存');
      setLoading(false);
    }, 1000);
  };

  const handleSystemSettingsSave = (values: SystemSettings) => {
    setLoading(true);
    // 模拟 API 调用
    setTimeout(() => {
      message.success('系统设置已保存');
      setLoading(false);
    }, 1000);
  };

  const handlePasswordChange = (values: { oldPassword: string; newPassword: string; confirmPassword: string }) => {
    if (values.newPassword !== values.confirmPassword) {
      message.error('两次输入的新密码不一致');
      return;
    }

    setLoading(true);
    // 模拟 API 调用
    setTimeout(() => {
      message.success('密码已修改');
      setLoading(false);
    }, 1000);
  };

  return (
    <Card title="设置">
      <Tabs defaultActiveKey="user">
        <TabPane tab="用户设置" key="user">
          <Form
            form={userForm}
            layout="vertical"
            onFinish={handleUserSettingsSave}
          >
            <Form.Item
              name="email"
              label="邮箱"
              rules={[
                { required: true, message: '请输入邮箱' },
                { type: 'email', message: '请输入有效的邮箱地址' }
              ]}
            >
              <Input placeholder="请输入邮箱" />
            </Form.Item>
            <Form.Item
              name="notifications"
              label="接收通知"
              valuePropName="checked"
            >
              <Switch />
            </Form.Item>
            <Form.Item>
              <Button type="primary" htmlType="submit" loading={loading}>
                保存
              </Button>
            </Form.Item>
          </Form>
        </TabPane>
        <TabPane tab="系统设置" key="system">
          <Form
            form={systemForm}
            layout="vertical"
            onFinish={handleSystemSettingsSave}
          >
            <Form.Item
              name="serverAddress"
              label="STUN 服务器地址"
              rules={[{ required: true, message: '请输入 STUN 服务器地址' }]}
            >
              <Input placeholder="请输入 STUN 服务器地址" />
            </Form.Item>
            <Form.Item
              name="serverPort"
              label="STUN 服务器端口"
              rules={[{ required: true, message: '请输入 STUN 服务器端口' }]}
            >
              <InputNumber min={1} max={65535} style={{ width: '100%' }} />
            </Form.Item>
            <Form.Item
              name="enableUPnP"
              label="启用 UPnP"
              valuePropName="checked"
            >
              <Switch />
            </Form.Item>
            <Form.Item
              name="enableRelay"
              label="启用中继"
              valuePropName="checked"
            >
              <Switch />
            </Form.Item>
            <Form.Item
              name="maxConnections"
              label="最大连接数"
              rules={[{ required: true, message: '请输入最大连接数' }]}
            >
              <InputNumber min={1} max={1000} style={{ width: '100%' }} />
            </Form.Item>
            <Form.Item>
              <Button type="primary" htmlType="submit" loading={loading}>
                保存
              </Button>
            </Form.Item>
          </Form>
        </TabPane>
        <TabPane tab="修改密码" key="password">
          <Form
            layout="vertical"
            onFinish={handlePasswordChange}
          >
            <Form.Item
              name="oldPassword"
              label="当前密码"
              rules={[{ required: true, message: '请输入当前密码' }]}
            >
              <Input.Password placeholder="请输入当前密码" />
            </Form.Item>
            <Form.Item
              name="newPassword"
              label="新密码"
              rules={[
                { required: true, message: '请输入新密码' },
                { min: 6, message: '密码至少6个字符' }
              ]}
            >
              <Input.Password placeholder="请输入新密码" />
            </Form.Item>
            <Form.Item
              name="confirmPassword"
              label="确认新密码"
              rules={[
                { required: true, message: '请确认新密码' },
                ({ getFieldValue }) => ({
                  validator(_, value) {
                    if (!value || getFieldValue('newPassword') === value) {
                      return Promise.resolve();
                    }
                    return Promise.reject(new Error('两次输入的密码不一致'));
                  },
                }),
              ]}
            >
              <Input.Password placeholder="请确认新密码" />
            </Form.Item>
            <Form.Item>
              <Button type="primary" htmlType="submit" loading={loading}>
                修改密码
              </Button>
            </Form.Item>
          </Form>
        </TabPane>
      </Tabs>
    </Card>
  );
};

export default Settings;
