import axios from 'axios';
import { CreateTask } from '../entities/models/CreateTask';
import { Task } from '../entities/models/Task';
import { getToken } from '../utils/authHelpers';

const BASE_URL = `${import.meta.env.VITE_BACKEND_URL}/tasks`;

export const createTask = async (task: CreateTask) => {
	try {
		const token = getToken();
		const url = `${BASE_URL}/create`;
		const response = await axios.post(url, task, {
			headers: {
				Authorization: `Bearer ${token}`,
			},
		});
		return response;
	} catch (error) {
		if (axios.isAxiosError(error)) {
			console.error('Error response:', error.response?.data);
			throw new Error(
				error.response?.data?.message || 'An error occurred while creating the task.'
			);
		} else {
			console.error('Unexpected error:', error);
			throw new Error('An unexpected error occurred.');
		}
	}
};

export const getTasksByProjectId = async (projectId: string): Promise<Task[]> => {
	try {
		const token = getToken();
		const response = await axios.get(`${BASE_URL}/${projectId}/tasks`, {
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

export type TaskStatusCountType = {
	PENDING: number;
	IN_PROGRESS: number;
	FINISHED: number;
};

export const getTaskCountsByProjectId = async (
	projectId: string
): Promise<{ totalTasks: number; taskStatusCounts: TaskStatusCountType }> => {
	try {
		const token = getToken();
		const response = await axios.get(`${BASE_URL}/${projectId}/tasks-count`, {
			headers: {
				Authorization: `Bearer ${token}`,
			},
		});
		return response.data;
	} catch (error) {
		if (axios.isAxiosError(error)) {
			console.error('Error fetching task counts:', error.response?.data);
			throw new Error(
				error.response?.data?.message || 'An error occurred while fetching task counts.'
			);
		} else {
			console.error('Unexpected error:', error);
			throw new Error('An unexpected error occurred.');
		}
	}
};

export const assignMemberToTask = async (taskId: string, memberId: string) => {
	try {
		const token = getToken();
		const response = await axios.put(
			`${BASE_URL}/${taskId}/assign/${memberId}`,
			{},
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

export const toggleTaskStatus = async (taskId: string, memberId: string) => {
	try {
		const token = getToken();
		const response = await axios.put(
			`${BASE_URL}/${taskId}/member/${memberId}/toggle-status`,
			{},
			{
				headers: {
					Authorization: `Bearer ${token}`,
				},
			}
		);
		return response.data;
	} catch (error) {
		if (axios.isAxiosError(error)) {
			console.error('Error toggling task status:', error.response?.data);
			throw new Error(
				error.response?.data?.message || 'An error occurred while toggling the task status.'
			);
		} else {
			console.error('Unexpected error:', error);
			throw new Error('An unexpected error occurred.');
		}
	}
};

export const removeMemberFromTask = async (taskId: string) => {
	try {
		const token = getToken();
		const url = `${BASE_URL}/${taskId}/remove-member`;
		const response = await axios.put(
			url,
			{},
			{
				headers: {
					Authorization: `Bearer ${token}`,
				},
			}
		);
		return response.data;
	} catch (error) {
		if (axios.isAxiosError(error)) {
			console.error('Error removing member from task:', error.response?.data);
			throw new Error(
				error.response?.data?.message || 'An error occurred while removing the member.'
			);
		} else {
			console.error('Unexpected error:', error);
			throw new Error('An unexpected error occurred.');
		}
	}
};
