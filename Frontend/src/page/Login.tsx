import React from 'react';
import { Form, Input, Button, message } from 'antd';
import { loginUser } from '../services/userService';

interface LoginFormValues {
	username: string;
	password: string;
}

const LoginPage: React.FC = () => {
	const onFinish = async (values: LoginFormValues) => {
		try {
			const { username, password } = values;
			const data = await loginUser(username, password);

			const token = data.token;
			if (token) {
				localStorage.setItem('token', token);
				message.success('Login successful!');
				setTimeout(() => {
					window.location.href = '/';
				}, 1000); 
			}
		} catch (error) {
			message.error('Login failed. Please try again.');
		}
	};

	return (
		<div style={{ maxWidth: '400px', margin: '100px auto' }}>
			<h2 style={{ textAlign: 'center' }}>Login</h2>
			<Form name="login" initialValues={{ remember: true }} onFinish={onFinish}>
				<Form.Item
					name="username"
					rules={[{ required: true, message: 'Please input your username!' }]}
				>
					<Input placeholder="Username" />
				</Form.Item>

				<Form.Item
					name="password"
					rules={[{ required: true, message: 'Please input your password!' }]}
				>
					<Input.Password placeholder="Password" />
				</Form.Item>
				<Form.Item>
					<Button type="primary" htmlType="submit" block>
						Log in
					</Button>
				</Form.Item>
			</Form>
		</div>
	);
};

export default LoginPage;
