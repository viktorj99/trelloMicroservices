import axios from "axios";

const BASE_URL = `${import.meta.env.VITE_BACKEND_URL}/notifications`;

// NotifyMembers function to notify added members
export const notifyMembers = async (project_name: string, user_ids: string[]) => {
    try {
        const url = `${BASE_URL}/project/add`;

        console.log('Sending:', { project_name, user_ids });

        // Send the project name and member IDs to the backend
        const response = await axios.post(url, {
            project_name,
            user_ids
        });

        return response.data;
    } catch (error) {
        if (axios.isAxiosError(error)) {
            console.error('Error response:', error.response?.data);
            throw new Error(
                error.response?.data?.message || 'An error occurred while notifying members.'
            );
        } else {
            console.error('Unexpected error:', error);
            throw new Error('An unexpected error occurred.');
        }
    }
};