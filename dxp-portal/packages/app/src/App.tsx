import { createApp } from '@backstage/frontend-defaults';
import catalogPlugin from '@backstage/plugin-catalog/alpha';
import { navModule } from './modules/nav';
import argocdPlugin from '@roadiehq/backstage-plugin-argo-cd/alpha';

export default createApp({
  features: [
    catalogPlugin,
    navModule,
    argocdPlugin,
  ],
});
