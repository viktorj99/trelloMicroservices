import axios from 'axios';
import { RegistrationUser } from '../entities/models/RegistrationUser';
import { getToken } from '../utils/authHelpers';

const BASE_URL = `${import.meta.env.VITE_BACKEND_URL}/users`;

export const postData = async (user: RegistrationUser, captchaToken: string | null) => {
	try {
		console.log('User data:', user);
		console.log('Captcha token:', captchaToken);
		const url = `${BASE_URL}/register`;
		const response = await axios.post(url, user, {
			headers: {
				captcha_token: captchaToken || '',
			},
		});
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

export const loginUser = async (
	username: string,
	password: string,
	captchaToken: string | null
) => {
	try {
		const url = `${BASE_URL}/login`;
		const response = await axios.post(url, { username, password, captchaToken });
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

export const sendResetPasswordEmail = async (email: string) => {
	try {
		const url = `${BASE_URL}/forgot-password`;
		const response = await axios.post(url, { email });

		return response;
	} catch (error) {
		if (axios.isAxiosError(error)) {
			console.error('Error response:', error.response?.data);
			throw new Error(
				error.response?.data || 'Failed to send reset password email. Please try again.'
			);
		} else {
			console.error('Unexpected error:', error);
			throw new Error('An unexpected error occurred. Please try again.');
		}
	}
};

export const changeForgotPassword = async (code: string, newPassword: string) => {
	try {
		const url = `${BASE_URL}/forgot-password/change`;
		const response = await axios.post(url, { code, newPassword });
		return response;
	} catch (error) {
		if (axios.isAxiosError(error)) {
			console.error('Error response:', error.response?.data);
			throw new Error(error.response?.data || 'Failed to reset password.');
		} else {
			console.error('Unexpected error:', error);
			throw new Error('An unexpected error occurred.');
		}
	}
};

export const requestMagicLink = async (email: string) => {
	try {
		const url = `${BASE_URL}/magic-link/request`;
		const response = await axios.post(url, { email });
		return response.data;
	} catch (error) {
		if (axios.isAxiosError(error)) {
			console.error('Error response:', error.response?.data);
			throw new Error(error.response?.data || 'Failed to send magic link. Please try again.');
		} else {
			throw new Error('An unexpected error occurred. Please try again.');
		}
	}
};

export const magicLogin = async (token: string) => {
	try {
		const url = `${BASE_URL}/magic-link/login`;
		const response = await axios.post(url, { token });
		return response.data;
	} catch (error) {
		if (axios.isAxiosError(error)) {
			console.error('Error response:', error.response?.data);
			throw new Error(error.response?.data || 'Invalid or expired magic link token.');
		} else {
			throw new Error('An unexpected error occurred.');
		}
	}
};

export const changePassword = async (newPassword: string) => {
	try {
		const url = `${BASE_URL}/change-password`;
		const token = getToken();
		const response = await axios.post(
			url,
			{ newPassword },
			{
				headers: {
					Authorization: `Bearer ${token}`,
				},
			}
		);
		return response.data;
	} catch (error) {
		if (axios.isAxiosError(error)) {
			throw new Error(error.response?.data || 'Failed to change password.');
		} else {
			console.error('Unexpected error:', error);
			throw new Error('An unexpected error occurred.');
		}
	}
};

export const deleteUser = async (id: number) => {
	try {
		const url = `${BASE_URL}/delete/${id}`;
		const token = getToken();
		const response = await axios.delete(url, {
			headers: {
				Authorization: `Bearer ${token}`,
			},
		});
		return response.data;
	} catch (error) {
		if (axios.isAxiosError(error)) {
			throw new Error(error.response?.data || 'Failed to delete users.');
		} else {
			console.error('Unexpected error:', error);
			throw new Error('An unexpected error occurred.');
		}
	}
};
