import { Form, Input, DatePicker, Button, Select, notification } from 'antd';
import { CreateProject } from '../entities/models/CreateProject';
import { User } from '../entities/models/User'; 
import { Role } from '../entities/models/Role';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { createProject } from '../services/projectService';

const { Option } = Select;

const CreateProjectForm: React.FC = () => {
  const [form] = Form.useForm();

  const queryClient = useQueryClient();
	const mutation = useMutation({
		mutationFn: createProject,
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: ['project'] });
			notification.success({
				message: 'Success',
				description: 'Projects created successfully!',
			});
			form.resetFields();
		},
		onError: (error) => {
			notification.error({
				message: 'Error',
				description: `Project creation failed: ${(error as Error).message}`,
			});
		},
	});

  

  const onFinish = (values: CreateProject) => {
    const selectedManager:User = {
      username: values.manager?.username as string,
      role: Role.Manager,
    };

    const selectedMembers: User[] = (values.members || []).map((memberUsername: string) => ({
      username: memberUsername,
      role: Role.Member, 
    }));

    mutation.mutate({
      ...values,
      manager: selectedManager,
      members: selectedMembers
    });    
  };

  

  const users: User[] = [
    {
      username: "john_doe",
      role: Role.Member,
    },
    {
      username: "jane_smith",
      role: Role.Member,
    },
    {
      username: "aliceUZemljiCuda",
      role: Role.Member,
    },
  ];

 

  return (
    <Form
      form={form}
      layout="vertical"
      onFinish={onFinish}
      initialValues={{ role: 'member' }} 
    >
      <Form.Item
        label="Project Name"
        name="name"
        rules={[{ required: true, message: 'Please input the project name!' }]}
      >
        <Input placeholder="Enter project name" />
      </Form.Item>

      <Form.Item
        label="Expected End Date"
        name="expectedEndDate"
        rules={[{ required: true, message: 'Please select the expected end date!' }]}
      >
        <DatePicker style={{ width: '100%' }} />
      </Form.Item>

      <Form.Item
        label="Minimum Members"
        name="minMembers"
        rules={[{ required: true, message: 'Please input the minimum number of members!' }]}
      >
        <Input type="number" placeholder="Enter minimum members" />
      </Form.Item>

      <Form.Item
        label="Maximum Members"
        name="maxMembers"
        rules={[{ required: true, message: 'Please input the maximum number of members!' }]}
      >
        <Input type="number" placeholder="Enter maximum members" />
      </Form.Item>

      <Form.Item
        label="Manager"
        name="manager"
        rules={[{ required: true, message: 'Please select a manager!' }]}
      >
        <Select placeholder="Select a manager">
          {users.map((user) => (
            <Option key={user.username} value={user.username}>
              {user.username}
            </Option>
          ))}
        </Select>
      </Form.Item>

      <Form.Item
        label="Members"
        name="members"
      >
        <Select mode="multiple" placeholder="Select members">
          {users.map((user) => (
            <Option key={user.username} value={user.username}>
              {user.username}
            </Option>
          ))}
        </Select>
      </Form.Item>

      <Form.Item>
        <Button type="primary" htmlType="submit">
          Create Project
        </Button>
      </Form.Item>
    </Form>
  );
};

export default CreateProjectForm;
