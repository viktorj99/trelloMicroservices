import React from 'react';
import { Form, Input, Button, Checkbox } from 'antd';

interface LoginFormValues {
	username: string;
	password: string;
	remember: boolean;
}

const LoginPage: React.FC = () => {
	const onFinish = (values: LoginFormValues) => {
		console.log('Success:', values);
	};

	return (
		<div style={{ maxWidth: '400px', margin: '100px auto' }}>
			<h2 style={{ textAlign: 'center' }}>Login</h2>
			<Form name='login' initialValues={{ remember: true }} onFinish={onFinish}>
				<Form.Item
					name='username'
					rules={[{ required: true, message: 'Please input your username!' }]}
				>
					<Input placeholder='Username' />
				</Form.Item>

				<Form.Item
					name='password'
					rules={[{ required: true, message: 'Please input your password!' }]}
				>
					<Input.Password placeholder='Password' />
				</Form.Item>

				<Form.Item name='remember' valuePropName='checked'>
					<Checkbox>Remember me</Checkbox>
				</Form.Item>

				<Form.Item>
					<Button type='primary' htmlType='submit' block>
						Log in
					</Button>
				</Form.Item>
			</Form>
		</div>
	);
};

export default LoginPage;
