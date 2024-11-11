import axios from 'axios';
import { RegistrationUser } from '../entities/models/RegistrationUser';
import { getToken } from '../utils/authHelpers';

export const postData = async (user: RegistrationUser) => {
	try {
		const url = `${import.meta.env.VITE_USER_BACKEND_URL}/register`;

		const response = await axios.post(url, user);
		return response;
	} catch (error) {
		console.log(error);
		if (axios.isAxiosError(error)) {
			// Log the error and throw a new error with a detailed message
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

export const getAllUserMembers = async () => {
	try {
		const url = `${import.meta.env.VITE_USER_BACKEND_URL}/users/members`;
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
	  // Make the request to the Go backend for verification
	  const url = `${import.meta.env.VITE_USER_BACKEND_URL}/verification`; // Assuming the backend has a /verify-code endpoint
	  const response = await axios.post(url, { code });
  
	  // Return the response from the backend
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
