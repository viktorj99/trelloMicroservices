import axios from 'axios';
import { getToken } from '../utils/authHelpers';

const BASE_URL = `${import.meta.env.VITE_BACKEND_URL}/workflow`;

export const addTaskDependency = async (taskId: string, dependencyId: string) => {
	console.log('taskId', taskId);
	console.log('dependencyId', dependencyId);
	try {
		const token = getToken();
		const response = await axios.post(
			`${BASE_URL}/dependencies`,
			{
				taskId: taskId,
				dependentId: dependencyId,
			},
			{
				headers: {
					Authorization: `Bearer ${token}`,
				},
			}
		);
		return response.data;
	} catch (error) {
		if (axios.isAxiosError(error)) {
			console.error('Error assigning member to task:', error.response?.data);
			throw new Error(error.response?.data || 'Circular dependency detected.');
		} else {
			console.error('Unexpected error:', error);
			throw new Error('An unexpected error occurred.');
		}
	}
};
