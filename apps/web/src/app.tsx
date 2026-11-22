// App de desarrollo: aquí se montan los microcomponentes para probarlos con
// los mocks. No se publica. Por ahora solo muestra que la base funciona.
export function App() {
  return (
    <main className="mx-auto max-w-3xl p-6">
      <h1 className="text-2xl font-semibold">Productos · app de desarrollo</h1>
      <p className="mt-2 text-sm text-gray-600">
        Base lista: Vite, React, Tailwind, tipos del contrato y mocks MSW. Los
        microcomponentes se montan aquí en los próximos PRs.
      </p>
    </main>
  )
}
