import axios from 'axios';
import { DTOCreateProject } from '../entities/models/CreateProject';
import { User } from '../entities/models/User';

export const createProject = async (user: DTOCreateProject) => {
	try {
		const url = `${import.meta.env.VITE_PROJECT_BACKEND_URL}/project/create`;

		const response = await axios.post(url, user);
		return response;
	} catch (error) {
		if (axios.isAxiosError(error)) {
			console.error('Error response:', error.response?.data);
			throw new Error(
				error.response?.data || 'An error occurred while creating the project.'
			);
		} else {
			console.error('Unexpected error:', error);
			throw new Error('An unexpected error occurred.');
		}
	}
};

export const getProject = async (id: string) => {
	try {
		const url = `${import.meta.env.VITE_PROJECT_BACKEND_URL}/project/${id}`;

		const response = await axios.get(url);
		return response.data;
	} catch (error) {
		if (axios.isAxiosError(error)) {
			console.error('Error response:', error.response?.data);
			throw new Error(
				error.response?.data || 'An error occurred while creating the project.'
			);
		} else {
			console.error('Unexpected error:', error);
			throw new Error('An unexpected error occurred.');
		}
	}
};

export const handleDeleteMember = async (username: string, members: User[], id: string) => {
	const newMembers = members.filter((member) => member.username != username);

	const payload = {
		members: newMembers,
	};
	console.log(newMembers);
	try {
		await axios.put(`${import.meta.env.VITE_PROJECT_BACKEND_URL}/project/${id}`, payload);
	} catch (error) {
		if (axios.isAxiosError(error)) {
			console.error('Error response:', error.response?.data);
			throw new Error(
				error.response?.data?.message || 'An error occurred while creating the project.'
			);
		} else {
			console.error('Unexpected error:', error);
			throw new Error('An unexpected error occurred.');
		}
	}
};

export const addMember = async (newMember: User, members: User[], id: string) => {
	const updatedMembers = [...members, newMember];

	const payload = {
		members: updatedMembers,
	};
	try {
		const url = `${import.meta.env.VITE_PROJECT_BACKEND_URL}/project/${id}`;
		await axios.put(url, payload);
	} catch (error) {
		if (axios.isAxiosError(error)) {
			console.error('Error response:', error.response?.data);
			throw new Error(
				error.response?.data?.message || 'An error occurred while adding the member.'
			);
		} else {
			console.error('Unexpected error:', error);
			throw new Error('An unexpected error occurred.');
		}
	}
}

export const getAllProjects = async () => {
	try {
		const url = `${import.meta.env.VITE_PROJECT_BACKEND_URL}/projects`;
		const response = await axios.get(url);
		return response.data;
	} catch (error) {
		if (axios.isAxiosError(error)) {
			console.error('Error response:', error.response?.data);
		} else {
			console.error('Unexpected error:', error);
		}
	}
};