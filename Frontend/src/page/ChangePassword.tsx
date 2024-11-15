import React, { useState } from 'react';
import { Form, Input, Button, notification } from 'antd';
import { changePassword } from '../services/userService'; // Backend API function

const ChangePasswordPage: React.FC = () => {
	const [code, setCode] = useState<string>('');
	const [newPassword, setNewPassword] = useState<string>('');
	const [confirmPassword, setConfirmPassword] = useState<string>('');
	const [loading, setLoading] = useState(false);

	const onFinish = async () => {
		if (newPassword !== confirmPassword) {
			notification.error({
				message: 'Error',
				description: 'Passwords do not match.',
			});
			return;
		}

		setLoading(true);
		try {
			await changePassword(code, newPassword);
			notification.success({
				message: 'Success',
				description: 'Password has been successfully changed. Please log in with your new password.',
			});
			setTimeout(() => {
				window.location.href = '/login'; 
			}, 1000);
		} catch (error) {
			notification.error({
				message: 'Error',
				description: 'Failed to reset password. Please try again.',
			});
		} finally {
			setLoading(false);
		}
	};

	return (
		<div style={{ maxWidth: 400, margin: '0 auto', padding: '2rem' }}>
			<h2>Change Password</h2>
			<Form onFinish={onFinish} layout="vertical">
				<Form.Item
					label="Verification Code"
					name="code"
					rules={[{ required: true, message: 'Please input the verification code!' }]}
				>
					<Input
						value={code}
						onChange={(e) => setCode(e.target.value)}
						placeholder="Enter the code"
					/>
				</Form.Item>

				<Form.Item
					label="New Password"
					name="newPassword"
					rules={[{ required: true, message: 'Please input your new password!' }]}
				>
					<Input.Password
						value={newPassword}
						onChange={(e) => setNewPassword(e.target.value)}
						placeholder="Enter new password"
					/>
				</Form.Item>

				<Form.Item
					label="Confirm Password"
					name="confirmPassword"
					rules={[{ required: true, message: 'Please confirm your new password!' }]}
				>
					<Input.Password
						value={confirmPassword}
						onChange={(e) => setConfirmPassword(e.target.value)}
						placeholder="Confirm new password"
					/>
				</Form.Item>

				<Form.Item>
					<Button type="primary" htmlType="submit" loading={loading} block>
						Submit
					</Button>
				</Form.Item>
			</Form>
		</div>
	);
};

export default ChangePasswordPage;
