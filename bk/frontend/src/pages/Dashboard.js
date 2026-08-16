import React from 'react';
import {
  Grid,
  Card,
  CardContent,
  Typography,
  Box,
  LinearProgress,
  Alert,
} from '@mui/material';
import {
  Storage as RAGIcon,
  SmartToy as AgentIcon,
  CheckCircle as ConnectedIcon,
  SmartToy as AssistantIcon,
  Chat as DialogueIcon,
  Article as ArticleIcon,
} from '@mui/icons-material';
import { useQuery } from 'react-query';
import api from '../services/api';

const Dashboard = ({ embedded = false }) => {
  const { data: status, isLoading, error } = useQuery('status', api.getStatus, { staleTime: 5 * 60 * 1000 }); // Cache for 5 minutes
  const { data: collections } = useQuery('collections', api.getRAGCollections, { staleTime: 5 * 60 * 1000 });
  const { data: agents } = useQuery('agents', api.getAgents, { staleTime: 5 * 60 * 1000 });
  const { data: assistants } = useQuery('assistants', api.getAssistants, { staleTime: 5 * 60 * 1000 });
  // assistants already fetched above
  const { data: dialogues } = useQuery('dialogues', api.getDialogues, { staleTime: 5 * 60 * 1000 });
  const { data: articleProfiles } = useQuery('articles', api.getArticles, { staleTime: 5 * 60 * 1000 });
  const { data: conversations } = useQuery('conversations', api.getConversations, { staleTime: 5 * 60 * 1000 });

  if (isLoading) {
    return (
      <Box sx={{ height: embedded ? '100%' : 'auto', minHeight: embedded ? 0 : 'auto', width: '100%' }}>
        <LinearProgress />
      </Box>
    );
  }

  if (error) {
    return (
      <Box sx={{ height: embedded ? '100%' : 'auto', minHeight: embedded ? 0 : 'auto' }}>
        <Alert severity="error">Failed to load dashboard data</Alert>
      </Box>
    );
  }

  const stats = [
    {
      title: 'RAG Collections',
      value: collections?.length || 0,
      icon: <RAGIcon />,
      color: 'primary',
    },
    {
      title: 'Active Agents',
      value: agents?.length || 0,
      icon: <AgentIcon />,
      color: 'secondary',
    },
    {
      title: 'Assistants',
      value: assistants?.length || 0,
      icon: <AssistantIcon />,
      color: 'info',
    },
    {
      title: 'Dialogues',
      value: dialogues?.length || 0,
      icon: <DialogueIcon />,
      color: 'warning',
    },
    {
      title: 'Articles',
      value: articleProfiles?.length || 0,
      icon: <ArticleIcon />,
      color: 'info',
    },
    {
      title: 'Conversations',
      value: conversations?.length || 0,
      icon: <DialogueIcon />,
      color: 'primary',
    },
  ];

  return (
    <Box sx={{ p: embedded ? 0 : 1, pb: embedded ? 0.5 : 1, minHeight: embedded ? 0 : 'auto', boxSizing: 'border-box' }}>
      {!embedded && (
        <Typography variant="h5" gutterBottom sx={{ mb: 1.25 }}>
          Dashboard
        </Typography>
      )}

      {/* System Status */}
      <Card sx={{ mb: 2.5, boxShadow: 1 }}>
        <CardContent sx={{ py: 1.5, px: 1.75 }}>
          <Typography variant="subtitle2" gutterBottom sx={{ display: 'flex', alignItems: 'center', gap: 0.75 }}>
            <ConnectedIcon color="success" sx={{ fontSize: 20 }} />
            System Status
          </Typography>
          <Typography variant="body2" color="text.secondary" sx={{ mt: 0.5 }}>
            {status?.available_models?.length || 0} models available
          </Typography>
        </CardContent>
      </Card>

      {/* Statistics */}
      <Typography variant="subtitle1" gutterBottom sx={{ mb: 1, fontWeight: 600 }}>
        Overview
      </Typography>
      <Grid container spacing={1.25} sx={{ mb: 2.5 }}>
        {stats.map((stat) => (
          <Grid item xs={12} sm={6} md={4} lg={3} key={stat.title}>
            <Card sx={{ height: '100%', minHeight: 96, boxShadow: 1, transition: 'transform 0.2s', '&:hover': { transform: 'translateY(-2px)' } }}>
              <CardContent sx={{ py: 1.25, px: 1.35 }}>
                <Box sx={{ display: 'flex', alignItems: 'center', gap: 1.25 }}>
                  <Box sx={{ color: `${stat.color}.main`, '& .MuiSvgIcon-root': { fontSize: 28 } }}>
                    {stat.icon}
                  </Box>
                  <Box sx={{ flex: 1, minWidth: 0 }}>
                    <Typography variant="h6" component="div" sx={{ fontWeight: 700, fontSize: '1.1rem', lineHeight: 1.2 }}>
                      {stat.value}
                    </Typography>
                    <Typography variant="caption" color="text.secondary" sx={{ display: 'block', mt: 0.35, lineHeight: 1.28 }}>
                      {stat.title}
                    </Typography>
                  </Box>
                </Box>
              </CardContent>
            </Card>
          </Grid>
        ))}
      </Grid>

      {/* Recent Activity */}
      <Typography variant="subtitle1" gutterBottom sx={{ mb: 1, fontWeight: 600 }}>
        Recent Activity
      </Typography>
      <Grid container spacing={1.25}>
        <Grid item xs={12} sm={6} md={6} lg={3}>
          <Card sx={{ boxShadow: 1, height: '100%' }}>
            <CardContent sx={{ py: 1.35, px: 1.35 }}>
              <Typography variant="subtitle2" gutterBottom sx={{ display: 'flex', alignItems: 'center', gap: 0.65 }}>
                <RAGIcon color="primary" sx={{ fontSize: 20 }} />
                Recent RAG Collections
              </Typography>
              <Box sx={{ maxHeight: 176, overflowY: 'auto', '&::-webkit-scrollbar': { width: '8px' }, '&::-webkit-scrollbar-track': { bgcolor: 'background.default', borderRadius: '4px' }, '&::-webkit-scrollbar-thumb': { bgcolor: 'primary.main', bgcolorOpacity: 0.5, borderRadius: '4px', '&:hover': { bgcolor: 'primary.light' } } }}>
                {collections?.slice(0, 5).map((collection) => (
                  <Box key={collection.name} sx={{ mb: 1, p: 0.85, borderRadius: 1, bgcolor: 'background.paper', border: '1px solid', borderColor: 'primary.main', borderOpacity: 0.3 }}>
                    <Typography variant="body2" sx={{ fontWeight: 600, fontSize: '0.8rem' }}>
                      {collection.name}
                    </Typography>
                    <Typography variant="caption" color="text.secondary">
                      {collection.count} documents
                    </Typography>
                  </Box>
                ))}
                {(!collections || collections.length === 0) && (
                  <Typography variant="body2" color="text.secondary" sx={{ textAlign: 'center', py: 1.5 }}>
                    No collections yet
                  </Typography>
                )}
              </Box>
            </CardContent>
          </Card>
        </Grid>

        <Grid item xs={12} sm={6} md={6} lg={3}>
          <Card sx={{ boxShadow: 1, height: '100%' }}>
            <CardContent sx={{ py: 1.35, px: 1.35 }}>
              <Typography variant="subtitle2" gutterBottom sx={{ display: 'flex', alignItems: 'center', gap: 0.65 }}>
                <AgentIcon color="secondary" sx={{ fontSize: 20 }} />
                Active Agents
              </Typography>
              <Box sx={{ maxHeight: 176, overflowY: 'auto', '&::-webkit-scrollbar': { width: '8px' }, '&::-webkit-scrollbar-track': { bgcolor: 'background.default', borderRadius: '4px' }, '&::-webkit-scrollbar-thumb': { bgcolor: 'primary.main', bgcolorOpacity: 0.5, borderRadius: '4px', '&:hover': { bgcolor: 'primary.light' } } }}>
                {agents?.slice(0, 5).map((agent) => (
                  <Box key={agent.id} sx={{ mb: 1, p: 0.85, borderRadius: 1, bgcolor: 'background.paper', border: '1px solid', borderColor: 'primary.main', borderOpacity: 0.3 }}>
                    <Typography variant="body2" sx={{ fontWeight: 600, fontSize: '0.8rem' }}>
                      {agent.name}
                    </Typography>
                    <Typography variant="caption" color="text.secondary">
                      {agent.model_name}
                    </Typography>
                  </Box>
                ))}
                {(!agents || agents.length === 0) && (
                  <Typography variant="body2" color="text.secondary" sx={{ textAlign: 'center', py: 1.5 }}>
                    No active agents
                  </Typography>
                )}
              </Box>
            </CardContent>
          </Card>
        </Grid>

        <Grid item xs={12} sm={6} md={6} lg={3}>
          <Card sx={{ boxShadow: 1, height: '100%' }}>
            <CardContent sx={{ py: 1.35, px: 1.35 }}>
              <Typography variant="subtitle2" gutterBottom sx={{ display: 'flex', alignItems: 'center', gap: 0.65 }}>
                <AssistantIcon color="info" sx={{ fontSize: 20 }} />
                Recent Assistants
              </Typography>
              <Box sx={{ maxHeight: 176, overflowY: 'auto', '&::-webkit-scrollbar': { width: '8px' }, '&::-webkit-scrollbar-track': { bgcolor: 'background.default', borderRadius: '4px' }, '&::-webkit-scrollbar-thumb': { bgcolor: 'primary.main', bgcolorOpacity: 0.5, borderRadius: '4px', '&:hover': { bgcolor: 'primary.light' } } }}>
                {assistants?.slice(0, 5).map((assistant) => (
                  <Box key={assistant.id} sx={{ mb: 1, p: 0.85, borderRadius: 1, bgcolor: 'background.paper', border: '1px solid', borderColor: 'primary.main', borderOpacity: 0.3 }}>
                    <Typography variant="body2" sx={{ fontWeight: 600, fontSize: '0.8rem' }}>
                      {assistant.name}
                    </Typography>
                    <Typography variant="caption" color="text.secondary">
                      {assistant.llm_provider || 'N/A'}
                    </Typography>
                  </Box>
                ))}
                {(!assistants || assistants.length === 0) && (
                  <Typography variant="body2" color="text.secondary" sx={{ textAlign: 'center', py: 1.5 }}>
                    No assistants yet
                  </Typography>
                )}
              </Box>
            </CardContent>
          </Card>
        </Grid>

        <Grid item xs={12} sm={6} md={6} lg={3}>
          <Card sx={{ boxShadow: 1, height: '100%' }}>
            <CardContent sx={{ py: 1.35, px: 1.35 }}>
              <Typography variant="subtitle2" gutterBottom sx={{ display: 'flex', alignItems: 'center', gap: 0.65 }}>
                <DialogueIcon color="warning" sx={{ fontSize: 20 }} />
                Recent Dialogues
              </Typography>
              <Box sx={{ maxHeight: 176, overflowY: 'auto', '&::-webkit-scrollbar': { width: '8px' }, '&::-webkit-scrollbar-track': { bgcolor: 'background.default', borderRadius: '4px' }, '&::-webkit-scrollbar-thumb': { bgcolor: 'primary.main', bgcolorOpacity: 0.5, borderRadius: '4px', '&:hover': { bgcolor: 'primary.light' } } }}>
                {dialogues?.slice(0, 5).map((dialogue) => (
                  <Box key={dialogue.id} sx={{ mb: 1, p: 0.85, borderRadius: 1, bgcolor: 'background.paper', border: '1px solid', borderColor: 'primary.main', borderOpacity: 0.3 }}>
                    <Typography variant="body2" sx={{ fontWeight: 600, fontSize: '0.8rem' }}>
                      {dialogue.name}
                    </Typography>
                    <Typography variant="caption" color="text.secondary">
                      {dialogue.description || 'No description'}
                    </Typography>
                  </Box>
                ))}
                {(!dialogues || dialogues.length === 0) && (
                  <Typography variant="body2" color="text.secondary" sx={{ textAlign: 'center', py: 1.5 }}>
                    No dialogues yet
                  </Typography>
                )}
              </Box>
            </CardContent>
          </Card>
        </Grid>
      </Grid>
    </Box>
  );
};

export default Dashboard; 