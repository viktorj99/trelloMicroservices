import React, { useState } from 'react';
import { Form, Input, Button, Select, notification } from 'antd';
import { Role } from '../entities/models/Role';
import { RegistrationUser } from '../entities/models/RegistrationUser';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { postData } from '../services/userService';
import DOMPurify from 'dompurify';
import ReCAPTCHA from 'react-google-recaptcha';

const RegistrationPage: React.FC = () => {
	const [form] = Form.useForm<RegistrationUser>();
	const queryClient = useQueryClient();
	const [captchaToken, setCaptchaToken] = useState<string | null>(null); 
	
	const mutation = useMutation({
		mutationFn: (user: RegistrationUser) => postData(user, captchaToken),
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: ['users'] });
			notification.success({
				message: 'Success',
				description: 'Registration successful, Check your Email.',
			});
			form.resetFields();
			setTimeout(() => {
				window.location.href = '/verification';
			}, 1000);
		},
		onError: (error) => {
			notification.error({
				message: 'Error',
				description: ` ${(error as Error).message}`,
			});
		},
	});

    // Form submission with sanitization
    const onFinish = (values: RegistrationUser) => {
        const sanitizedValues = {
            ...values,
            first_name: DOMPurify.sanitize(values.first_name),
            last_name: DOMPurify.sanitize(values.last_name),
            email: DOMPurify.sanitize(values.email),
            username: DOMPurify.sanitize(values.username),
        };
        mutation.mutate(sanitizedValues);
    };

	const onCaptchaChange = (token: string | null) => {
		setCaptchaToken(token);
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
                    rules={[
                        { required: true, message: 'Please input your first name!' },
                        { pattern: /^[A-Za-z]+$/, message: 'First name can only contain letters!' },
                    ]}
                >
                    <Input />
                </Form.Item>

                <Form.Item
                    name='last_name'
                    label='Last Name'
                    rules={[
                        { required: true, message: 'Please input your last name!' },
                        { pattern: /^[A-Za-z]+$/, message: 'Last name can only contain letters!' },
                    ]}
                >
                    <Input />
                </Form.Item>

                <Form.Item
                    name='email'
                    label='Email'
                    rules={[
                        { required: true, message: 'Please input your email!' },
                        { type: 'email', message: 'Please enter a valid email address!' },
                    ]}
                >
                    <Input />
                </Form.Item>

                <Form.Item
                    name='username'
                    label='Username'
                    rules={[
                        { required: true, message: 'Please input your username!' },
                        { min: 3, message: 'Username must be at least 3 characters!' },
                        { max: 20, message: 'Username cannot exceed 20 characters!' },
                        {
                            pattern: /^[A-Za-z0-9_]+$/,
                            message: 'Username can only contain letters, numbers, and underscores!',
                        },
                    ]}
                >
                    <Input />
                </Form.Item>

                <Form.Item
                    name='password'
                    label='Password'
                    rules={[
                        { required: true, message: 'Please input your password!' },
                        { min: 6, message: 'Password must be at least 6 characters!' },
                    ]}
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
				<div style={{ textAlign: 'center', marginBottom: '10px' }}>
					<ReCAPTCHA
						sitekey={import.meta.env.VITE_RECAPTCHA_SITE_KEY}
						onChange={onCaptchaChange}
					/>
				</div>

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
