import axios from 'axios';
import { getToken, getTokenData } from '../utils/authHelpers';

const BASE_URL = `${import.meta.env.VITE_BACKEND_URL}/activities`;

export const getActivityHistory = async () => {
	const token = getToken();
	const tokenData = getTokenData();

	if (!tokenData || !tokenData.id) {
		throw new Error('User not authenticated');
	}

	const response = await axios.get(`${BASE_URL}/user/${tokenData.id}`, {
		headers: {
			Authorization: `Bearer ${token}`,
		},
	});
	return response.data;
};