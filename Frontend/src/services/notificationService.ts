import axios from "axios";

const BASE_URL = `${import.meta.env.VITE_BACKEND_URL}/notifications`;

// NotifyMembers function to notify added members
export const notifyMembers = async (projectName: string, memberIds: string[]) => {
    try {
        const url = `${BASE_URL}/project/add`;

        // Send the project name and member IDs to the backend
        const response = await axios.post(url, {
            projectName,
            memberIds
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