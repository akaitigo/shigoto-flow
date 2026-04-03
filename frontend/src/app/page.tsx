import Link from "next/link";

export default function Home() {
  return (
    <div className="min-h-screen bg-gray-50">
      <header className="border-b bg-white">
        <div className="mx-auto flex max-w-6xl items-center justify-between px-6 py-4">
          <h1 className="text-xl font-bold text-gray-900">Shigoto-Flow</h1>
          <nav className="flex gap-6">
            <Link href="/reports" className="text-gray-600 hover:text-gray-900">
              レポート
            </Link>
            <Link
              href="/settings"
              className="text-gray-600 hover:text-gray-900"
            >
              設定
            </Link>
          </nav>
        </div>
      </header>

      <main className="mx-auto max-w-6xl px-6 py-12">
        <div className="text-center">
          <h2 className="text-3xl font-bold text-gray-900">
            書かない日報、始めよう
          </h2>
          <p className="mt-4 text-lg text-gray-600">
            Googleカレンダー・Slack・GitHub・Gmailから活動データを自動集約。
            <br />
            日報テンプレートに自動流し込みして、確認・修正するだけ。
          </p>
        </div>

        <div className="mt-12 grid gap-8 sm:grid-cols-3">
          <FeatureCard
            title="自動集約"
            description="4つのデータソースから当日の活動を自動で収集します"
          />
          <FeatureCard
            title="日報生成"
            description="集約データをテンプレートに流し込み、確認して送信するだけ"
          />
          <FeatureCard
            title="週報・月報"
            description="日報からAIが週報・月報を自動要約。手間なく上長に報告"
          />
        </div>
      </main>
    </div>
  );
}

function FeatureCard({
  title,
  description,
}: {
  title: string;
  description: string;
}) {
  return (
    <div className="rounded-lg border border-gray-200 bg-white p-6">
      <h3 className="text-lg font-semibold text-gray-900">{title}</h3>
      <p className="mt-2 text-sm text-gray-600">{description}</p>
    </div>
  );
}
