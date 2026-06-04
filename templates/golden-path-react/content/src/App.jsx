function App() {
  return (
    <div style={{ fontFamily: 'sans-serif', padding: '2rem' }}>
      <h1>${{ values.name }}</h1>
      <p>${{ values.description }}</p>
      <p style={{ color: '#666' }}>Déployé sur DxP</p>
    </div>
  );
}
export default App;
