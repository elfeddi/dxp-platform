import React from 'react';
import { createFrontendPlugin } from '@backstage/frontend-plugin-api';
import { EntityContentBlueprint } from '@backstage/plugin-catalog-react/alpha';

const dxpOverviewContent = EntityContentBlueprint.make({
  name: 'DxpOverview',
  params: {
    path: '/dxp-overview',
    title: 'DxP Overview',
    loader: () =>
      import('./DxpOverviewPage').then(m =>
        React.createElement(m.DxpOverviewPage),
      ),
  },
});

export const dxpOverviewPlugin = createFrontendPlugin({
  id: 'dxp-overview',
  extensions: [dxpOverviewContent],
});
