import { StyledEngineProvider } from '@mui/material/styles';
import { Layout } from './Layout';
import { createRoot } from 'react-dom/client';

const container = document.getElementById('commits');
const root = createRoot(container!);
root.render(<StyledEngineProvider injectFirst><Layout /></StyledEngineProvider>);