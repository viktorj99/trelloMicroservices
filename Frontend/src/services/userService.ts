import axios from 'axios';
import { RegistrationUser } from '../entities/models/RegistrationUser';
import { getToken } from '../utils/authHelpers';

const BASE_URL = `${import.meta.env.VITE_BACKEND_URL}/users`;

export const postData = async (user: RegistrationUser) => {
    try {
        const url = `${BASE_URL}/register`;
        const response = await axios.post(url, user);
        return response;
    } catch (error) {
        if (axios.isAxiosError(error)) {
            console.error('Error response:', error.response?.data);
            throw new Error(error.response?.data || 'Registration failed. Please try again.');
        } else {
            console.error('Unexpected error:', error);
            throw new Error('An unexpected error occurred. Please try again.');
        }
    }
};

export const loginUser = async (username: string, password: string) => {
    try {
        const url = `${BASE_URL}/login`;
        const response = await axios.post(url, { username, password });
        return response.data;
    } catch (error) {
        if (axios.isAxiosError(error)) {
            console.error('Error response:', error.response?.data);
            throw new Error(error.response?.data?.message || 'Login failed');
        } else {
            console.error('Unexpected error:', error);
            throw new Error('Unexpected error occurred');
        }
    }
};

export const getAllUserMembers = async () => {
    try {
        const url = `${BASE_URL}/members`;
        const token = getToken();
        const response = await axios.get(url, {
            headers: {
                Authorization: `Bearer ${token}`,
            },
        });
        return response.data;
    } catch (error) {
        if (axios.isAxiosError(error)) {
            console.error('Error response:', error.response?.data);
            throw new Error(error.response?.data || 'Cannot get all user members');
        } else {
            console.error('Unexpected error:', error);
            throw new Error('Unexpected error occurred');
        }
    }
};

export const verifyCode = async (code: string) => {
    try {
        const url = `${BASE_URL}/verification`;
        const response = await axios.post(url, { code });
        return response;
    } catch (error) {
        if (axios.isAxiosError(error)) {
            console.error('Error response:', error.response?.data);
            throw new Error(error.response?.data || 'Verification failed. Please try again.');
        } else {
            console.error('Unexpected error:', error);
            throw new Error('An unexpected error occurred. Please try again.');
        }
    }
};
