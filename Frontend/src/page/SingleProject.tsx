import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useParams } from 'react-router-dom';
import { addMember, getProject, handleDeleteMember } from '../services/projectService';
import { Button, Form, Input, Modal, notification, Select, Table } from 'antd';
import { User } from '../entities/models/User';
import { useState, useEffect } from 'react';
import { Role } from '../entities/models/Role';
import { createTask, getTasksByProjectId, assignMemberToTask, toggleTaskStatus } from '../services/taskService';
import { Task } from '../entities/models/Task';
import { getTokenData } from '../utils/authHelpers';
import { getAllUserMembers } from '../services/userService';
import DOMPurify from 'dompurify';
const { Option } = Select;

const SingleProject = () => {
  const { id } = useParams<{ id: string }>();
  console.log('Project ID from URL:', id);
  const [isModalVisible, setIsModalVisible] = useState(false);
  const [isTaskModalVisible, setIsTaskModalVisible] = useState(false);
  const [selectedUser, setSelectedUser] = useState<string | null>(null);
  const [taskForm] = Form.useForm();
  const [tasks, setTasks] = useState<Task[]>([]);

  const [userRole, setUserRole] = useState<Role | null>(null); 
  const [userId, setUserId] = useState<string | null>(null);

  const { data: project, isLoading, error } = useQuery({
    queryKey: ['project', id],
    queryFn: () => getProject(id!),
  });

  const { data: users, isLoading: usersLoading } = useQuery<User[]>({
    queryKey: ['users'],
    queryFn: getAllUserMembers,
  });

  useEffect(() => {
    if (id) {
      getTasksByProjectId(id)
        .then((tasks) => setTasks(tasks))  
        .catch(console.error);
    }

    const tokenPayload = getTokenData(); 
    setUserRole(tokenPayload.role);
    setUserId(tokenPayload.id);
  }, [id]);

  const queryClient = useQueryClient();

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

  const handleCreateTask = (values: any) => {
    // Sanitizacija unosa pomoću DOMPurify
    const sanitizedValues = {
      ...values,
      title: DOMPurify.sanitize(values.title),
      description: DOMPurify.sanitize(values.description),
    };

    taskMutation.mutate(sanitizedValues);
  };

  if (isLoading || usersLoading) {
    return <p>Loading project data...</p>;
  }

  if (error) {
    return <p>Error: {(error as Error).message}</p>;
  }

  if (!project) {
    return <p>No project data available.</p>;
  }

  return (
    <div>
      <h1>{project.name}</h1>
      <p>Expected End Date: {project.expectedEndDate}</p>
      <p>Max Members: {project.maxMembers}</p>
      <p>Min Members: {project.minMembers}</p>

      {userRole === Role.Manager && (
        <>
          <Button type="primary" onClick={() => setIsModalVisible(true)}>
            Add Member
          </Button>
          <Button
            type="primary"
            onClick={() => setIsTaskModalVisible(true)}
            style={{ marginLeft: 16 }}
          >
            Create Task
          </Button>
        </>
      )}

      {/* Modal za kreiranje zadatka */}
      <Modal
        title="Create Task"
        open={isTaskModalVisible}
        onCancel={() => setIsTaskModalVisible(false)}
        footer={null}
      >
        <Form form={taskForm} layout="vertical" onFinish={handleCreateTask}>
          <Form.Item
            label="Task Title"
            name="title"
            rules={[
              { required: true, message: 'Please enter the task title!' },
              { min: 5, message: 'Title must be at least 5 characters long.' },
              { max: 100, message: 'Title cannot exceed 100 characters.' },
            ]}
          >
            <Input placeholder="Enter task title" />
          </Form.Item>

          <Form.Item
            label="Description"
            name="description"
            rules={[
              { required: true, message: 'Please enter the description!' },
              { min: 10, message: 'Description must be at least 10 characters long.' },
            ]}
          >
            <Input.TextArea placeholder="Enter task description" />
          </Form.Item>

          <Form.Item name="status" initialValue="PENDING" hidden>
            <Input type="hidden" />
          </Form.Item>

          <Form.Item label="Project ID" name="project" initialValue={id} hidden />

          <Form.Item>
            <Button type="primary" htmlType="submit">
              Create Task
            </Button>
          </Form.Item>
        </Form>
      </Modal>

      {/* Tasks Table */}
      <h2>Tasks</h2>
      <Table dataSource={tasks} rowKey="id">
        <Table.Column title="Task Title" dataIndex="title" />
        <Table.Column title="Description" dataIndex="description" />
        <Table.Column title="Status" dataIndex="status" />
      </Table>
    </div>
  );
};

export default SingleProject;
