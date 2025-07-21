import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { getTokenData, removeToken } from '../utils/authHelpers';
import { getAllProjectsWithManagerId, getAllProjectsWithUserId } from '../services/projectService';
import { DTOCreateProject } from '../entities/models/CreateProject';
import { notification } from 'antd';
import { deleteUser } from '../services/userService';
import { useNavigate } from 'react-router-dom';

const Profile = () => {
	const queryClient = useQueryClient();
	const navigate = useNavigate();
	const user = getTokenData();
	console.log('User:', user);
	const { data: projects } = useQuery<DTOCreateProject[]>({
		queryKey: ['projectUsers'],
		queryFn: () => {
			if (user.role === 'Member') {
				return getAllProjectsWithUserId(user.id);
			} else {
				return getAllProjectsWithManagerId(user.id);
			}
		},
	});

	const mutation = useMutation({
		mutationFn: (id: number) => deleteUser(id),
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: ['project'] });
			notification.success({
				message: 'Success',
				description: 'Profile deleted successfully!',
			});
		},
		onError: (error) => {
			notification.error({
				message: 'Error',
				description: `${(error as Error).message}`,
			});
		},
	});

	const handleDeleteUser = () => {
		const today = new Date();
		today.setHours(0, 0, 0, 0);
		let forDelete = false;
		projects?.forEach((project) => {
			const projectEndDate = new Date(project.expectedEndDate + 'T00:00:00Z');
			if (projectEndDate > today) {
				notification.error({
					message: 'Error',
					description: `You cant delete profile because you have active projects`,
				});
			} else {
				forDelete = true;
			}
		});
		if (forDelete || projects === null) {
			mutation.mutate(user.id);
			removeToken();
			navigate('/login');
		}
	};
	return <button onClick={handleDeleteUser}>Izbrisi profil</button>;
};

export default Profile;
