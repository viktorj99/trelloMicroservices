import axios from 'axios';
import { DTOCreateProject } from '../entities/models/CreateProject';
import { User } from '../entities/models/User';
import { getToken } from '../utils/authHelpers';

const BASE_URL = `${import.meta.env.VITE_BACKEND_URL}/projects`;

export const createProject = async (user: DTOCreateProject) => {
    try {
        const url = `${BASE_URL}/create`;
        const token = getToken();
        const response = await axios.post(url, user, {
            headers: {
                Authorization: `Bearer ${token}`,
            },
        });
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
			const url = `${BASE_URL}/${id}`;
			const token = getToken();
			console.log('Requesting project with URL:', url);

			const response = await axios.get(url, {
					headers: {
							Authorization: `Bearer ${token}`,
					},
			});

			console.log('Fetched project data:', response.data);
			return response.data;
	} catch (error) {
			if (axios.isAxiosError(error)) {
					console.error('Error response:', error.response?.data);
					throw new Error(
							error.response?.data || 'An error occurred while fetching the project.'
					);
			} else {
					console.error('Unexpected error:', error);
					throw new Error('An unexpected error occurred.');
			}
	}
};


export const handleDeleteMember = async (username: string, members: User[], id: string) => {
    const newMembers = members.filter((member) => member.username !== username);

    const payload = {
        members: newMembers,
    };

    try {
        const token = getToken();
        const url = `${BASE_URL}/${id}/update`;
        await axios.put(url, payload, {
            headers: {
                Authorization: `Bearer ${token}`,
            },
        });
    } catch (error) {
        if (axios.isAxiosError(error)) {
            console.error('Error response:', error.response?.data);
            throw new Error(
                error.response?.data?.message || 'An error occurred while updating the project.'
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
        const token = getToken();
        const url = `${BASE_URL}/${id}/update`;
        await axios.put(url, payload, {
            headers: {
                Authorization: `Bearer ${token}`,
            },
        });
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
};

export const getAllProjects = async () => {
    try {
        const token = getToken();
        const url = `${BASE_URL}/`;
        const response = await axios.get(url, {
            headers: {
                Authorization: `Bearer ${token}`,
            },
        });
        return response.data;
    } catch (error) {
        if (axios.isAxiosError(error)) {
            console.error('Error response:', error.response?.data);
        } else {
            console.error('Unexpected error:', error);
        }
    }
};