import axios from 'axios';
import { CreateTask } from '../entities/models/CreateTask';

export const createTask = async (task: CreateTask) => {
    try {
        const url = `${import.meta.env.VITE_TASK_BACKEND_URL}/task/create`;

        const response = await axios.post(url, task);
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