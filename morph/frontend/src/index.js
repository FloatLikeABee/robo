import React from 'react';
import ReactDOM from 'react-dom/client';
import { RouterProvider } from 'react-router-dom';
import { LocalizationProvider } from '@mui/x-date-pickers/LocalizationProvider';
import { AdapterDayjs } from '@mui/x-date-pickers/AdapterDayjs';
import './index.css';
import { appRouter } from './appRouter';
import { ConfirmProvider } from './components/ConfirmDialog';

const root = ReactDOM.createRoot(document.getElementById('root'));
root.render(
  <React.StrictMode>
    <LocalizationProvider dateAdapter={AdapterDayjs}>
      <ConfirmProvider>
        <RouterProvider router={appRouter} />
      </ConfirmProvider>
    </LocalizationProvider>
  </React.StrictMode>,
);
