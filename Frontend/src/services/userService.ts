import axios from 'axios';
import { RegistrationUser } from '../entities/models/RegistrationUser';

export const postData = async (user: RegistrationUser) => {
	try {
		const url = `${import.meta.env.VITE_USER_BACKEND_URL}/register`;

		const response = await axios.post(url, user);
		return response;
	} catch (error) {
		if (axios.isAxiosError(error)) {
			console.error('Error response:', error.response?.data);
		} else {
			console.error('Unexpected error:', error);
		}
	}
};

export const loginUser = async (username: string, password: string) => {
	try {
		const url = `${import.meta.env.VITE_USER_BACKEND_URL}/login`;
		const response = await axios.post(url, { username, password });
		return response.data; 
	} catch (error) {
		if (axios.isAxiosError(error)) {
			console.error('Error response:', error.response?.data);
			throw new Error(error.response?.data.message || 'Login failed');
		} else {
			console.error('Unexpected error:', error);
			throw new Error('Unexpected error occurred');
		}
	}
};