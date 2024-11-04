import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useParams } from 'react-router-dom';
import { addMember, getProject, handleDeleteMember } from '../services/projectService';
import { Button, Modal, notification, Select, Table } from 'antd';
import { User } from '../entities/models/User';
import { useState } from 'react';
import { Role } from '../entities/models/Role';
const { Option } = Select;

const users: User[] = [
	{
		username: 'john_doe',
		role: Role.Member,
	},
	{
		username: 'jane_smith',
		role: Role.Member,
	},
	{
		username: 'aliceUZemljiCuda',
		role: Role.Member,
	},
];

const SingleProject = () => {
	const { id } = useParams<{ id: string }>();
	const [isModalVisible, setIsModalVisible] = useState(false);
	const [selectedUser, setSelectedUser] = useState<string | null>(null);

	const { data: project } = useQuery({
		queryKey: ['project', id],
		queryFn: () => getProject(id!),
	});

	console.log(project);

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

	const addMutation = useMutation({
		mutationFn: (username: string) => {
			const newUser = users.find((user) => user.username === username);
			return addMember(newUser!, project.members, id!);
		},
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: ['project', id] });
			notification.success({
				message: 'Success',
				description: 'Member added successfully!',
			});
			setIsModalVisible(false);
		},
		onError: (error: unknown) => {
			notification.error({
				message: 'Error',
				description: `Member addition failed: ${(error as Error).message}`,
			});
		},
	});

	const showModal = () => {
		setIsModalVisible(true);
	};

	const handleOk = () => {
		if (selectedUser) {
			addMutation.mutate(selectedUser);
		} else {
			notification.error({
				message: 'Error',
				description: 'Please select a user to add.',
			});
		}
	};

	const handleCancel = () => {
		setIsModalVisible(false);
	};

	return (
		<div>
			<h1>{project?.name}</h1>
			<p>Expected End Date: {project?.expectedEndDate}</p>
			<p>Max Members: {project?.maxMembers}</p>
			<p>Min Members: {project?.minMembers}</p>

			<Button type='primary' onClick={showModal}>
				Add Member
			</Button>

			<Modal
				title='Add a Member'
				open={isModalVisible}
				onOk={handleOk}
				onCancel={handleCancel}
			>
				<Select
					placeholder='Select a user'
					style={{ width: '100%' }}
					onChange={(value) => setSelectedUser(value)}
				>
					{users.map((user) => (
						<Option key={user.username} value={user.username}>
							{user.username}
						</Option>
					))}
				</Select>
			</Modal>

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
