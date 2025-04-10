import { createSlice, createAsyncThunk } from '@reduxjs/toolkit';
import axios from 'axios';
import { API_URL } from '../../config';

interface Forward {
  id: string;
  protocol: 'tcp' | 'udp';
  srcPort: number;
  dstHost: string;
  dstPort: number;
  description: string;
  enabled: boolean;
  stats: {
    bytesSent: number;
    bytesReceived: number;
    connections: number;
    startTime: string;
  };
}

interface ForwardState {
  forwards: Forward[];
  currentForward: Forward | null;
  loading: boolean;
  error: string | null;
}

const initialState: ForwardState = {
  forwards: [],
  currentForward: null,
  loading: false,
  error: null,
};

export const fetchForwards = createAsyncThunk(
  'forward/fetchForwards',
  async (_, { rejectWithValue }) => {
    try {
      // TODO: 实现获取转发规则列表 API 调用
      // const response = await axios.get(`${API_URL}/forwards`);
      // return response.data;
      
      // 模拟转发规则列表
      return [
        {
          id: '1',
          protocol: 'tcp',
          srcPort: 23389,
          dstHost: 'localhost',
          dstPort: 3389,
          description: 'Remote Desktop',
          enabled: true,
          stats: {
            bytesSent: 1024000,
            bytesReceived: 2048000,
            connections: 5,
            startTime: new Date().toISOString(),
          },
        },
        {
          id: '2',
          protocol: 'tcp',
          srcPort: 2222,
          dstHost: '192.168.1.5',
          dstPort: 22,
          description: 'SSH Server',
          enabled: false,
          stats: {
            bytesSent: 0,
            bytesReceived: 0,
            connections: 0,
            startTime: new Date().toISOString(),
          },
        },
      ] as Forward[];
    } catch (error: any) {
      return rejectWithValue(error.response?.data?.message || '获取转发规则列表失败');
    }
  }
);

export const fetchForwardById = createAsyncThunk(
  'forward/fetchForwardById',
  async (id: string, { rejectWithValue }) => {
    try {
      // TODO: 实现获取转发规则详情 API 调用
      // const response = await axios.get(`${API_URL}/forwards/${id}`);
      // return response.data;
      
      // 模拟转发规则详情
      return {
        id,
        protocol: 'tcp',
        srcPort: 23389,
        dstHost: 'localhost',
        dstPort: 3389,
        description: 'Remote Desktop',
        enabled: true,
        stats: {
          bytesSent: 1024000,
          bytesReceived: 2048000,
          connections: 5,
          startTime: new Date().toISOString(),
        },
      } as Forward;
    } catch (error: any) {
      return rejectWithValue(error.response?.data?.message || '获取转发规则详情失败');
    }
  }
);

const forwardSlice = createSlice({
  name: 'forward',
  initialState,
  reducers: {
    clearForwardError: (state) => {
      state.error = null;
    },
  },
  extraReducers: (builder) => {
    builder
      .addCase(fetchForwards.pending, (state) => {
        state.loading = true;
        state.error = null;
      })
      .addCase(fetchForwards.fulfilled, (state, action) => {
        state.loading = false;
        state.forwards = action.payload;
      })
      .addCase(fetchForwards.rejected, (state, action) => {
        state.loading = false;
        state.error = action.payload as string;
      })
      .addCase(fetchForwardById.pending, (state) => {
        state.loading = true;
        state.error = null;
      })
      .addCase(fetchForwardById.fulfilled, (state, action) => {
        state.loading = false;
        state.currentForward = action.payload;
      })
      .addCase(fetchForwardById.rejected, (state, action) => {
        state.loading = false;
        state.error = action.payload as string;
      });
  },
});

export const { clearForwardError } = forwardSlice.actions;

export default forwardSlice.reducer;
