import React from 'react';
import { Routes, Route, Navigate } from 'react-router-dom';
import { Layout } from 'antd';
import Login from './pages/Login';
import Dashboard from './pages/Dashboard';
import Devices from './pages/Devices';
import DeviceDetail from './pages/DeviceDetail';
import Apps from './pages/Apps';
import AppDetail from './pages/AppDetail';
import Forwards from './pages/Forwards';
import ForwardDetail from './pages/ForwardDetail';
import Settings from './pages/Settings';
import NotFound from './pages/NotFound';
import AppLayout from './components/AppLayout';
import PrivateRoute from './components/PrivateRoute';
import './App.css';

const App: React.FC = () => {
  return (
    <Routes>
      <Route path="/login" element={<Login />} />
      <Route path="/" element={<PrivateRoute><AppLayout /></PrivateRoute>}>
        <Route index element={<Navigate to="/dashboard" replace />} />
        <Route path="dashboard" element={<Dashboard />} />
        <Route path="devices" element={<Devices />} />
        <Route path="devices/:id" element={<DeviceDetail />} />
        <Route path="apps" element={<Apps />} />
        <Route path="apps/:id" element={<AppDetail />} />
        <Route path="forwards" element={<Forwards />} />
        <Route path="forwards/:id" element={<ForwardDetail />} />
        <Route path="settings" element={<Settings />} />
      </Route>
      <Route path="*" element={<NotFound />} />
    </Routes>
  );
};

export default App;
