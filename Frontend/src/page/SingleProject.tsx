import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useParams } from 'react-router-dom';
import { getProject, handleDeleteMember } from '../services/projectService';
import { Button, notification, Table } from 'antd';
import { User } from '../entities/models/User';

const SingleProject = () => {
	const { id } = useParams<{ id: string }>();

	const { data: project } = useQuery({
		queryKey: ['project', id],
		queryFn: () => getProject(id!),
	});

	const queryClient = useQueryClient();

	const mutation = useMutation({
		mutationFn: (username: string) => handleDeleteMember(username, project.members, id!),
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: ['project', id] });
			notification.success({
				message: 'Success',
				description: 'Projects created successfully!',
			});
		},
		onError: (error: unknown) => {
			notification.error({
				message: 'Error',
				description: `Project creation failed: ${(error as Error).message}`,
			});
		},
	});

	console.log(project);
	return (
		<div>
			<h1>{project?.name}</h1>
			<p>Expected End Date: {project?.expectedEndDate}</p>
			<p>Max Members: {project?.maxMembers}</p>
			<p>Min Members: {project?.minMembers}</p>

			<Table dataSource={project?.members} rowKey='username'>
				<Table.Column title='Username' dataIndex='username' />
				<Table.Column title='Role' dataIndex='role' />
				<Table.Column
					title='Action'
					render={(_, record: User) => (
						<Button danger onClick={() => mutation.mutate(record.username)}>
							Delete
						</Button>
					)}
				/>
			</Table>
		</div>
	);
};

export default SingleProject;
