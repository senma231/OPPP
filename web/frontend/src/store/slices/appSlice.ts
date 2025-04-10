import { createSlice, createAsyncThunk } from '@reduxjs/toolkit';
import axios from 'axios';
import { API_URL } from '../../config';

interface App {
  id: string;
  name: string;
  protocol: 'tcp' | 'udp';
  srcPort: number;
  peerNode: string;
  dstPort: number;
  dstHost: string;
  status: 'running' | 'stopped' | 'error';
}

interface AppState {
  apps: App[];
  currentApp: App | null;
  loading: boolean;
  error: string | null;
}

const initialState: AppState = {
  apps: [],
  currentApp: null,
  loading: false,
  error: null,
};

export const fetchApps = createAsyncThunk(
  'app/fetchApps',
  async (_, { rejectWithValue }) => {
    try {
      // TODO: 实现获取应用列表 API 调用
      // const response = await axios.get(`${API_URL}/apps`);
      // return response.data;
      
      // 模拟应用列表
      return [
        {
          id: '1',
          name: 'Remote Desktop',
          protocol: 'tcp',
          srcPort: 23389,
          peerNode: 'Office-PC',
          dstPort: 3389,
          dstHost: 'localhost',
          status: 'running',
        },
        {
          id: '2',
          name: 'SSH Server',
          protocol: 'tcp',
          srcPort: 2222,
          peerNode: 'Office-PC',
          dstPort: 22,
          dstHost: '192.168.1.5',
          status: 'stopped',
        },
      ] as App[];
    } catch (error: any) {
      return rejectWithValue(error.response?.data?.message || '获取应用列表失败');
    }
  }
);

export const fetchAppById = createAsyncThunk(
  'app/fetchAppById',
  async (id: string, { rejectWithValue }) => {
    try {
      // TODO: 实现获取应用详情 API 调用
      // const response = await axios.get(`${API_URL}/apps/${id}`);
      // return response.data;
      
      // 模拟应用详情
      return {
        id,
        name: 'Remote Desktop',
        protocol: 'tcp',
        srcPort: 23389,
        peerNode: 'Office-PC',
        dstPort: 3389,
        dstHost: 'localhost',
        status: 'running',
      } as App;
    } catch (error: any) {
      return rejectWithValue(error.response?.data?.message || '获取应用详情失败');
    }
  }
);

const appSlice = createSlice({
  name: 'app',
  initialState,
  reducers: {
    clearAppError: (state) => {
      state.error = null;
    },
  },
  extraReducers: (builder) => {
    builder
      .addCase(fetchApps.pending, (state) => {
        state.loading = true;
        state.error = null;
      })
      .addCase(fetchApps.fulfilled, (state, action) => {
        state.loading = false;
        state.apps = action.payload;
      })
      .addCase(fetchApps.rejected, (state, action) => {
        state.loading = false;
        state.error = action.payload as string;
      })
      .addCase(fetchAppById.pending, (state) => {
        state.loading = true;
        state.error = null;
      })
      .addCase(fetchAppById.fulfilled, (state, action) => {
        state.loading = false;
        state.currentApp = action.payload;
      })
      .addCase(fetchAppById.rejected, (state, action) => {
        state.loading = false;
        state.error = action.payload as string;
      });
  },
});

export const { clearAppError } = appSlice.actions;

export default appSlice.reducer;
