import React, { useState } from 'react';
import { Badge, Drawer, List, Menu, Spin } from 'antd';
import { HomeOutlined, UserOutlined, SettingOutlined, BellOutlined } from '@ant-design/icons';
import { Link } from 'react-router-dom';
import { getTokenData } from '../../utils/authHelpers';
import { fetchNotifications } from '../../services/notificationService';

const Navbar: React.FC = () => {
	const tokenData = getTokenData();
	const isManager = tokenData?.role === 'Manager';
	const [drawerVisible, setDrawerVisible] = useState(false);
	const [notifications, setNotifications] = useState<any[]>([]);
	const [loading, setLoading] = useState(false);

	const loadNotifications = async () => {
		if (!tokenData || !tokenData.id) return;
		setLoading(true);
		try {
			const data = await fetchNotifications(tokenData.id);
			setNotifications(data);
		} catch (error) {
			console.error('Failed to fetch notifications:', error);
		} finally {
			setLoading(false);
		}
	};

	// Open the drawer and load notifications
	const openDrawer = () => {
		setDrawerVisible(true);
		loadNotifications();
	};

	// Close the drawer
	const closeDrawer = () => {
		setDrawerVisible(false);
	};


	return (
		<>
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
			{isManager && (
				<Menu.Item key='projectCreate' icon={<SettingOutlined />}>
					<Link to='/project/create'>Create Project</Link>
				</Menu.Item>
			)}
			<Menu.Item key='projects' icon={<SettingOutlined />}>
				<Link to='/projects'>Projects</Link>
			</Menu.Item>
			<Menu.Item key='notifications' icon={<BellOutlined />} onClick={openDrawer}>
					<Badge count={notifications.filter((n: { is_read: any; }) => !n.is_read).length}>
						Notifications
					</Badge>
				</Menu.Item>
		</Menu>
		{/* Notification Drawer */}
		
		<Drawer
		title='Notifications'
		placement='right'
		onClose={closeDrawer}
		open={drawerVisible}
		width={350}
		>
		{loading ? (
			<Spin />
		) : (
			<List
				dataSource={notifications}
				renderItem={item => (
					<List.Item>
						<List.Item.Meta title={item.message} description={new Date(item.created_at).toLocaleString()} />
					</List.Item>
				)}
			/>
		)}
		</Drawer>
	</>
	
	);
};

export default Navbar;
