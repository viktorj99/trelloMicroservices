import axios from 'axios';
import { getToken } from '../utils/authHelpers';
import { TaskWithDependencies } from '../entities/models/TaskWithDependencies';

const BASE_URL = `${import.meta.env.VITE_BACKEND_URL}/workflow`;

export const getTasksWithDependencies = async (): Promise<TaskWithDependencies[]> => {
	try {
		const token = getToken();
		const response = await axios.get(`${BASE_URL}/tasks/dependencies`, {
			headers: {
				Authorization: `Bearer ${token}`,
			},
		});
		return response.data;
	} catch (error) {
		if (axios.isAxiosError(error)) {
			console.error('Error fetching tasks:', error.response?.data);
			throw new Error(
				error.response?.data?.message || 'An error occurred while fetching tasks.'
			);
		} else {
			console.error('Unexpected error:', error);
			throw new Error('An unexpected error occurred.');
		}
	}
};

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
			throw new Error(
				error.response?.data?.message || 'An error occurred while assigning the member.'
			);
		} else {
			console.error('Unexpected error:', error);
			throw new Error('An unexpected error occurred.');
		}
	}
};
