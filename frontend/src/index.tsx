import { Layout } from './Layout';
import { createRoot } from 'react-dom/client';

const container = document.getElementById('commits');
const root = createRoot(container!);
root.render(<Layout />);