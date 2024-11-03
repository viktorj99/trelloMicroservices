import axios from 'axios';
import { DTOCreateProject } from '../entities/models/CreateProject';

export const createProject = async (user: DTOCreateProject) => {
	try {
		const url = `${import.meta.env.VITE_PROJECT_BACKEND_URL}/project/create`;

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
