import axios from 'axios';
import { RegistrationUser } from '../entities/models/RegistrationUser';

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
			throw new Error(
				error.response?.data || 'Registration failed. Please try again.'
			);
		} else {
			console.error('Unexpected error:', error);
			throw new Error('An unexpected error occurred. Please try again.');
		}
	}
};
