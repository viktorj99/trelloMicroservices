import React from 'react';
import { Menu } from 'antd';
import { HomeOutlined, UserOutlined, SettingOutlined } from '@ant-design/icons';
import { Link } from 'react-router-dom';

const Navbar: React.FC = () => {
	return (
		<Menu mode='horizontal' theme='dark' defaultSelectedKeys={['home']}>
			<Menu.Item key='home' icon={<HomeOutlined />}>
				<Link to='/'>Home</Link>
			</Menu.Item>
			<Menu.Item key='login' icon={<UserOutlined />}>
				<Link to='/login'>Login</Link>
			</Menu.Item>
			<Menu.Item key='registration' icon={<SettingOutlined />}>
				<Link to='/registration'>Registration</Link>
			</Menu.Item>
			<Menu.Item key='projectCreate' icon={<SettingOutlined />}>
				<Link to='/project/create'>Create Project</Link>
			</Menu.Item>
		</Menu>
	);
};

export default Navbar;
