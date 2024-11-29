import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useParams } from 'react-router-dom';
import { addMember, getProject, handleDeleteMember } from '../services/projectService';
import { Button, Form, Input, Modal, notification, Select, Table } from 'antd';
import { User } from '../entities/models/User';
import { useState } from 'react';
import { Role } from '../entities/models/Role';
import { createTask } from '../services/taskService';
import { notifyMembers } from '../services/notificationService';
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
	console.log('Project ID from URL:', id);
	const [isModalVisible, setIsModalVisible] = useState(false);
	const [isTaskModalVisible, setIsTaskModalVisible] = useState(false);
	const [selectedUser, setSelectedUser] = useState<string | null>(null);
	const [taskForm] = Form.useForm();

	// Hooks moraju biti na vrhu
	const { data: project, isLoading, error } = useQuery({
		queryKey: ['project', id],
		queryFn: () => getProject(id!),
	});

	const queryClient = useQueryClient();

	const mutation = useMutation({
		mutationFn: (username: string) => handleDeleteMember(username, project?.members || [], id!),
		onSuccess: async (data, variables) => {
			// Notify after successful deletion
			try {
				await notifyMembers(project.name, [variables], 1); // 1 corresponds to "project/remove"
				notification.success({
					message: 'Success',
					description: 'Member deleted and notified successfully!',
				});
			} catch (error) {
				notification.error({
					message: 'Notification Error',
					description: `Member was deleted but notification failed: ${(error as Error).message}`,
				});
			}
	
			queryClient.invalidateQueries({ queryKey: ['project', id] });
		},
		onError: (error: unknown) => {
			notification.error({
				message: 'Error',
				description: `Project update failed: ${(error as Error).message}`,
			});
		},
	});

	const addMutation = useMutation({
		mutationFn: (username: string) => {
			const newUser = users.find((user) => user.username === username);
			return addMember(newUser!, project?.members || [], id!);
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

	const taskMutation = useMutation({
		mutationFn: createTask,
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: ['tasks'] });
			notification.success({
				message: 'Success',
				description: 'Task created successfully!',
			});
			taskForm.resetFields();
			setIsTaskModalVisible(false);
		},
		onError: (error: unknown) => {
			notification.error({
				message: 'Error',
				description: `Task creation failed: ${(error as Error).message}`,
			});
		},
	});

	// Provera stanja učitavanja i grešaka
	if (isLoading) {
		return <p>Loading project data...</p>;
	}

	if (error) {
		return <p>Error: {(error as Error).message}</p>;
	}

	if (!project) {
		return <p>No project data available.</p>;
	}

	console.log('Fetched project:', project);

	const handleAddMember = () => {
		if (selectedUser) {
			addMutation.mutate(selectedUser);
		} else {
			notification.error({
				message: 'Error',
				description: 'Please select a user to add.',
			});
		}
	};

	const handleCreateTask = (values: any) => {
		taskMutation.mutate(values);
	};

	return (
		<div>
			<h1>{project.name}</h1>
			<p>Expected End Date: {project.expectedEndDate}</p>
			<p>Max Members: {project.maxMembers}</p>
			<p>Min Members: {project.minMembers}</p>

			<Button type='primary' onClick={() => setIsModalVisible(true)}>
				Add Member
			</Button>

			<Button type='primary' onClick={() => setIsTaskModalVisible(true)} style={{ marginLeft: 16 }}>
				Create Task
			</Button>

			{/* Modal for adding a member */}
			<Modal
				title='Add a Member'
				open={isModalVisible}
				onOk={handleAddMember}
				onCancel={() => setIsModalVisible(false)}
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

			{/* Modal for creating a task */}
			<Modal
				title='Create Task'
				open={isTaskModalVisible}
				onCancel={() => setIsTaskModalVisible(false)}
				footer={null}
			>
				<Form form={taskForm} layout='vertical' onFinish={handleCreateTask}>
					<Form.Item
						label='Task Title'
						name='title'
						rules={[{ required: true, message: 'Please enter the task title!' }]}
					>
						<Input placeholder='Enter task title' />
					</Form.Item>

					<Form.Item
						label='Description'
						name='description'
						rules={[{ required: true, message: 'Please enter the description!' }]}
					>
						<Input.TextArea placeholder='Enter task description' />
					</Form.Item>

					<Form.Item name='status' initialValue='PENDING' hidden>
						<Input type='hidden' />
					</Form.Item>

					<Form.Item label='Project ID' name='project' initialValue={id} hidden />

					<Form.Item>
						<Button type='primary' htmlType='submit'>
							Create Task
						</Button>
					</Form.Item>
				</Form>
			</Modal>

			<Table dataSource={project.members} rowKey='username'>
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

