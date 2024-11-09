import React from 'react';
import { Form, Input, Button, Select, notification } from 'antd';
import { Role } from '../entities/models/Role';
import { RegistrationUser } from '../entities/models/RegistrationUser';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { postData } from '../services/userService';

const RegistrationPage: React.FC = () => {
	const [form] = Form.useForm<RegistrationUser>();
	const queryClient = useQueryClient();
	const mutation = useMutation({
		mutationFn: postData,
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: ['users'] });
			notification.success({
				message: 'Success',
				description: 'Registration successful!',
			});
			form.resetFields();
		},
		onError: (error) => {
			notification.error({
				message: 'Error',
				description: ` ${(error as Error).message}`,
			});
		},
	});

	const onFinish = (values: RegistrationUser) => {
		mutation.mutate(values);
	};

	return (
		<div style={{ maxWidth: 400, margin: '0 auto', padding: '2rem' }}>
			<h2>Registration</h2>
			<Form
				form={form}
				name='register'
				onFinish={onFinish}
				scrollToFirstError
				layout='vertical'
			>
				<Form.Item
					name='first_name'
					label='First Name'
					rules={[{ required: true, message: 'Please input your first name!' }]}
				>
					<Input />
				</Form.Item>

				<Form.Item
					name='last_name'
					label='Last Name'
					rules={[{ required: true, message: 'Please input your last name!' }]}
				>
					<Input />
				</Form.Item>

				<Form.Item
					name='email'
					label='Email'
					rules={[
						{ type: 'email', message: 'The input is not valid E-mail!' },
						{ required: true, message: 'Please input your E-mail!' },
					]}
				>
					<Input />
				</Form.Item>

				<Form.Item
					name='username'
					label='Username'
					rules={[{ required: true, message: 'Please input your username!' }]}
				>
					<Input />
				</Form.Item>

				<Form.Item
					name='password'
					label='Password'
					rules={[{ required: true, message: 'Please input your password!' }]}
					hasFeedback
				>
					<Input.Password />
				</Form.Item>

				<Form.Item
					name='confirm'
					label='Confirm Password'
					dependencies={['password']}
					hasFeedback
					rules={[
						{ required: true, message: 'Please confirm your password!' },
						({ getFieldValue }) => ({
							validator(_, value) {
								if (!value || getFieldValue('password') === value) {
									return Promise.resolve();
								}
								return Promise.reject(new Error('The two passwords do not match!'));
							},
						}),
					]}
				>
					<Input.Password />
				</Form.Item>

				<Form.Item
					name='role'
					label='Role'
					rules={[{ required: true, message: 'Please select a role!' }]}
				>
					<Select placeholder='Select a role'>
						<Select.Option value={Role.Manager}>Manager</Select.Option>
						<Select.Option value={Role.Member}>Member</Select.Option>
					</Select>
				</Form.Item>

				<Form.Item>
					<Button type='primary' htmlType='submit'>
						Register
					</Button>
				</Form.Item>
			</Form>
		</div>
	);
};

export default RegistrationPage;
