import { createSlice, createAsyncThunk } from '@reduxjs/toolkit';
import axios from 'axios';
import { API_URL } from '../../config';

interface Device {
  id: string;
  name: string;
  status: 'online' | 'offline';
  natType: string;
  externalIP: string;
  lastSeen: string;
}

interface DeviceState {
  devices: Device[];
  currentDevice: Device | null;
  loading: boolean;
  error: string | null;
}

const initialState: DeviceState = {
  devices: [],
  currentDevice: null,
  loading: false,
  error: null,
};

export const fetchDevices = createAsyncThunk(
  'device/fetchDevices',
  async (_, { rejectWithValue }) => {
    try {
      // TODO: 实现获取设备列表 API 调用
      // const response = await axios.get(`${API_URL}/devices`);
      // return response.data;
      
      // 模拟设备列表
      return [
        {
          id: '1',
          name: 'Office-PC',
          status: 'online',
          natType: 'Port Restricted Cone NAT',
          externalIP: '203.0.113.1',
          lastSeen: new Date().toISOString(),
        },
        {
          id: '2',
          name: 'Home-PC',
          status: 'offline',
          natType: 'Symmetric NAT',
          externalIP: '203.0.113.2',
          lastSeen: new Date(Date.now() - 86400000).toISOString(),
        },
      ] as Device[];
    } catch (error: any) {
      return rejectWithValue(error.response?.data?.message || '获取设备列表失败');
    }
  }
);

export const fetchDeviceById = createAsyncThunk(
  'device/fetchDeviceById',
  async (id: string, { rejectWithValue }) => {
    try {
      // TODO: 实现获取设备详情 API 调用
      // const response = await axios.get(`${API_URL}/devices/${id}`);
      // return response.data;
      
      // 模拟设备详情
      return {
        id,
        name: 'Office-PC',
        status: 'online',
        natType: 'Port Restricted Cone NAT',
        externalIP: '203.0.113.1',
        lastSeen: new Date().toISOString(),
      } as Device;
    } catch (error: any) {
      return rejectWithValue(error.response?.data?.message || '获取设备详情失败');
    }
  }
);

const deviceSlice = createSlice({
  name: 'device',
  initialState,
  reducers: {
    clearDeviceError: (state) => {
      state.error = null;
    },
  },
  extraReducers: (builder) => {
    builder
      .addCase(fetchDevices.pending, (state) => {
        state.loading = true;
        state.error = null;
      })
      .addCase(fetchDevices.fulfilled, (state, action) => {
        state.loading = false;
        state.devices = action.payload;
      })
      .addCase(fetchDevices.rejected, (state, action) => {
        state.loading = false;
        state.error = action.payload as string;
      })
      .addCase(fetchDeviceById.pending, (state) => {
        state.loading = true;
        state.error = null;
      })
      .addCase(fetchDeviceById.fulfilled, (state, action) => {
        state.loading = false;
        state.currentDevice = action.payload;
      })
      .addCase(fetchDeviceById.rejected, (state, action) => {
        state.loading = false;
        state.error = action.payload as string;
      });
  },
});

export const { clearDeviceError } = deviceSlice.actions;

export default deviceSlice.reducer;
