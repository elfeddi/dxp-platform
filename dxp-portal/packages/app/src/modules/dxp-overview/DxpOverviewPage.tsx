import React, { useEffect, useState } from 'react';
import { useEntity } from '@backstage/plugin-catalog-react';

interface PipelineRun {
  name: string;
  status: string;
  start_time: string;
  duration: string;
}

interface Pod {
  name: string;
  status: string;
  ready: string;
  restarts: number;
  age: string;
}

interface ServiceOverview {
  name: string;
  namespace: string;
  pipeline: { runs: PipelineRun[]; total: number };
  deploy: {
    app_name: string;
    sync_status: string;
    health: string;
    revision: string;
    message?: string;
  };
  pods: Pod[];
}

const statusColor = (s: string): string => {
  if (['Succeeded', 'Synced', 'Healthy', 'Running'].includes(s)) return '#3B6D11';
  if (['Failed', 'Degraded', 'Error'].includes(s)) return '#A32D2D';
  if (['Progressing', 'Pending'].includes(s)) return '#185FA5';
  return '#5F5E5A';
};

const Badge = ({ label }: { label: string }) => (
  <span style={{
    fontSize: 11, fontWeight: 500, padding: '2px 8px',
    borderRadius: 20, background: '#F1EFE8',
    color: statusColor(label), marginLeft: 6,
  }}>{label}</span>
);

const Card = ({ title, children }: { title: string; children: React.ReactNode }) => (
  <div style={{
    border: '0.5px solid #D1D9E6', borderRadius: 10,
    padding: '16px 20px', marginBottom: 14, background: '#fff',
  }}>
    <div style={{ fontSize: 13, fontWeight: 500, color: '#1A2744', marginBottom: 12 }}>
      {title}
    </div>
    {children}
  </div>
);

const Row = ({ label, value, badge }: { label: string; value?: string; badge?: string }) => (
  <div style={{
    display: 'flex', justifyContent: 'space-between',
    padding: '5px 0', borderBottom: '0.5px solid #EEF1F7', fontSize: 13,
  }}>
    <span style={{ color: '#718096' }}>{label}</span>
    <span style={{ color: '#1C2333', fontWeight: 500 }}>
      {value}{badge && <Badge label={badge} />}
    </span>
  </div>
);

export const DxpOverviewPage = () => {
  const { entity } = useEntity();
  const serviceName = entity.metadata.name;
  const [data, setData] = useState<ServiceOverview | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    setLoading(true);
    fetch(`/api/proxy/dxp/api/dxp/service/${serviceName}`)
      .then(r => {
        if (!r.ok) throw new Error(`HTTP ${r.status}`);
        return r.json();
      })
      .then(d => { setData(d); setLoading(false); })
      .catch(e => { setError(e.message); setLoading(false); });
  }, [serviceName]);

  if (loading) {
    return (
      <div style={{ padding: 24, color: '#718096', fontSize: 14 }}>
        Chargement des données DxP...
      </div>
    );
  }

  if (error) {
    return (
      <div style={{ padding: 24, color: '#A32D2D', fontSize: 14 }}>
        Erreur : {error}
      </div>
    );
  }

  if (!data) return null;

  return (
    <div style={{ padding: 24, maxWidth: 860 }}>

      <Card title="Déploiement — ArgoCD">
        <Row label="Application" value={data.deploy.app_name} />
        <Row label="Sync" badge={data.deploy.sync_status || '—'} />
        <Row label="Santé" badge={data.deploy.health || '—'} />
        <Row label="Révision" value={data.deploy.revision || '—'} />
        {data.deploy.message && (
          <Row label="Message" value={data.deploy.message} />
        )}
      </Card>

      <Card title={`Pods — ${data.namespace}`}>
        {data.pods.length === 0 ? (
          <div style={{ color: '#718096', fontSize: 13 }}>Aucun pod</div>
        ) : (
          data.pods.map(p => (
            <div key={p.name} style={{
              display: 'flex', justifyContent: 'space-between',
              alignItems: 'center', padding: '5px 0',
              borderBottom: '0.5px solid #EEF1F7', fontSize: 13,
            }}>
              <span style={{ color: '#1C2333', fontFamily: 'monospace', fontSize: 12 }}>
                {p.name}
              </span>
              <span>
                <Badge label={p.status} />
                <span style={{ color: '#718096', marginLeft: 8 }}>{p.ready}</span>
                <span style={{ color: '#718096', marginLeft: 8 }}>{p.age}</span>
                {p.restarts > 0 && (
                  <span style={{ color: '#A32D2D', marginLeft: 8 }}>
                    {p.restarts} restarts
                  </span>
                )}
              </span>
            </div>
          ))
        )}
      </Card>

      <Card title={`Pipeline Tekton — ${data.pipeline.total} runs`}>
        {!data.pipeline.runs || data.pipeline.runs.length === 0 ? (
          <div style={{ color: '#718096', fontSize: 13 }}>Aucun pipeline run</div>
        ) : (
          data.pipeline.runs.map(r => (
            <div key={r.name} style={{
              display: 'flex', justifyContent: 'space-between',
              alignItems: 'center', padding: '5px 0',
              borderBottom: '0.5px solid #EEF1F7', fontSize: 13,
            }}>
              <span style={{ color: '#1C2333', fontFamily: 'monospace', fontSize: 12 }}>
                {r.name}
              </span>
              <span>
                <Badge label={r.status} />
                <span style={{ color: '#718096', marginLeft: 8 }}>{r.start_time}</span>
                <span style={{ color: '#718096', marginLeft: 8 }}>{r.duration}</span>
              </span>
            </div>
          ))
        )}
      </Card>

    </div>
  );
};
