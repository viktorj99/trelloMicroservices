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
