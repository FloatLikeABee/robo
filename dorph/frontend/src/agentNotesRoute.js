import React from 'react';
import { Navigate } from 'react-router-dom';

/** Signed-in /agent-notes lands in chat with the agent-notes modal open. */
export const agentNotesRoute = {
  path: 'agent-notes',
  element: <Navigate to="/?agent-notes=1" replace />,
};
