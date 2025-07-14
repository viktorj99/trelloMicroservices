import axios from 'axios';
import { getToken } from '../utils/authHelpers';

const BASE_URL = `${import.meta.env.VITE_BACKEND_URL}`;

export const uploadTaskDocument = (taskId: string, formData: FormData) => {
  const token = getToken();
  return axios.post(`${BASE_URL}/tasks/${taskId}/documents`, formData, {
    headers: {
      Authorization: `Bearer ${token}`,
    },
  });
};

export const getTaskDocuments = (taskId: string) => {
  const token = getToken();
  return axios
    .get(`${BASE_URL}/tasks/${taskId}/documents`, {
      headers: {
        Authorization: `Bearer ${token}`,
      },
    })
    .then((res) => res.data);
};

export const getDocumentDownloadUrl = (taskId: string, fileName: string): string => {
  return `${BASE_URL}/tasks/${taskId}/documents/${encodeURIComponent(fileName)}/download`;
};


export const deleteTaskDocument = (taskId: string, fileName: string) => {
  const token = getToken();
  return axios.delete(`${BASE_URL}/tasks/${taskId}/documents/${encodeURIComponent(fileName)}`, {
    headers: { Authorization: `Bearer ${token}` },
  });
};