import { configureStore } from '@reduxjs/toolkit';
import authReducer from './slices/authSlice';
import deviceReducer from './slices/deviceSlice';
import appReducer from './slices/appSlice';
import forwardReducer from './slices/forwardSlice';

export const store = configureStore({
  reducer: {
    auth: authReducer,
    device: deviceReducer,
    app: appReducer,
    forward: forwardReducer,
  },
});

export type RootState = ReturnType<typeof store.getState>;
export type AppDispatch = typeof store.dispatch;
