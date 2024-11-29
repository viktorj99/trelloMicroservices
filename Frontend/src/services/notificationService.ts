import axios from "axios";

const BASE_URL = `${import.meta.env.VITE_BACKEND_URL}/notifications`;

const getPathFromNumber = (num: number): string => {
    switch (num) {
        case 0:
            return "/project/add";
        case 1:
            return "/project/remove";
        case 2:
            return "/task/add";
        case 3:
            return "/task/remove";
        case 4:
            return "/task/status";
        default:
            throw new Error("Invalid notification type.");
    }
};


// NotifyMembers function to notify added members
export const notifyMembers = async (project_name: string, user_ids: string[], type: number) => {
    try {
        const path = getPathFromNumber(type);
        const url = `${BASE_URL}${path}`;

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

export const fetchNotifications = async (userId: string) => {

    console.log("Sending: " , {userId});

    try {
        const response = await axios.get(`${BASE_URL}/user`, {
            headers: { user_id: userId },
        });
        return response.data;
    } catch (error) {
        if (axios.isAxiosError(error)) {
            console.error('Error fetching notifications:', error.response?.data);
            throw new Error(
                error.response?.data?.message || 'Error fetching notifications.'
            );
        } else {
            console.error('Unexpected error:', error);
            throw new Error('An unexpected error occurred.');
        }
    }
};