"use client";

import { useCallback, useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { ApiError, authFetch } from "@/lib/api";

type Device = {
  id: string;
  name: string;
  description: string;
  status: string;
  location_id: number | null;
  location?: { id: number; name: string } | null;
  longitude: number | null;
  latitude: number | null;
  altitude: number | null;
  latest_health?: unknown;
  created_at: string;
  updated_at: string;
};

type DeviceList = {
  data: Device[];
  page: number;
  limit: number;
  total: number;
  total_page: number;
};

type Location = { id: number; name: string };

type StatusHist = {
  id: string;
  device_id: string;
  status: string;
  created_at: string;
};

type ApiKeyMeta = {
  id: number;
  device_id: string;
  masking: string;
  created_at: string;
  is_revoked: boolean;
};

const STATUSES = [
  "installed",
  "active",
  "disconnected",
  "maintenance",
  "error",
  "decommissioned",
];

const EMPTY_FORM = {
  id: "",
  name: "",
  description: "",
  status: "installed",
  location_id: "",
  longitude: "",
  latitude: "",
  altitude: "",
};

function toPayload(f: typeof EMPTY_FORM, isEdit: boolean) {
  const num = (v: string) => (v === "" ? null : Number(v));
  const body: Record<string, unknown> = {
    name: f.name,
    description: f.description,
    status: f.status,
    location_id: f.location_id === "" ? null : Number(f.location_id),
    longitude: num(f.longitude),
    latitude: num(f.latitude),
    altitude: num(f.altitude),
  };
  if (!isEdit) body.id = f.id;
  return body;
}

export default function DevicesPage() {
  const router = useRouter();
  const [list, setList] = useState<DeviceList | null>(null);
  const [locations, setLocations] = useState<Location[]>([]);
  const [page, setPage] = useState(1);
  const [q, setQ] = useState("");
  const [qInput, setQInput] = useState("");
  const [status, setStatus] = useState("");
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [showForm, setShowForm] = useState(false);
  const [editing, setEditing] = useState<Device | null>(null);
  const [form, setForm] = useState(EMPTY_FORM);
  const [formError, setFormError] = useState("");
  const [rawKey, setRawKey] = useState("");
  const [detail, setDetail] = useState<Device | null>(null);
  const [history, setHistory] = useState<StatusHist[]>([]);
  const [creds, setCreds] = useState<ApiKeyMeta[]>([]);
  const [rotated, setRotated] = useState<{ id: string; rawKey: string } | null>(
    null,
  );

  function handleAuthError(err: unknown) {
    if (err instanceof ApiError && err.status === 401) {
      router.replace("/login");
      return true;
    }
    return false;
  }

  const load = useCallback(async () => {
    setLoading(true);
    setError("");
    try {
      const params = new URLSearchParams({
        page: String(page),
        limit: "10",
      });
      if (q) params.set("q", q);
      if (status) params.set("status", status);
      const data = await authFetch<DeviceList>(
        `/api/v1/devices?${params.toString()}`,
      );
      setList(data);
    } catch (err) {
      if (!handleAuthError(err))
        setError(err instanceof Error ? err.message : "Gagal memuat data");
    } finally {
      setLoading(false);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [page, q, status]);

  useEffect(() => {
    load();
  }, [load]);

  useEffect(() => {
    authFetch<{ data: Location[] }>(`/api/v1/locations?limit=100`)
      .then((d) => setLocations(d.data ?? []))
      .catch(() => {});
  }, []);

  function openCreate() {
    setEditing(null);
    setForm(EMPTY_FORM);
    setFormError("");
    setRawKey("");
    setShowForm(true);
  }

  function openEdit(dev: Device) {
    setEditing(dev);
    setForm({
      id: dev.id,
      name: dev.name,
      description: dev.description ?? "",
      status: dev.status,
      location_id: dev.location_id != null ? String(dev.location_id) : "",
      longitude: dev.longitude != null ? String(dev.longitude) : "",
      latitude: dev.latitude != null ? String(dev.latitude) : "",
      altitude: dev.altitude != null ? String(dev.altitude) : "",
    });
    setFormError("");
    setRawKey("");
    setShowForm(true);
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setFormError("");
    try {
      if (editing) {
        await authFetch(`/api/v1/devices/${editing.id}`, {
          method: "PATCH",
          body: JSON.stringify(toPayload(form, true)),
        });
        setShowForm(false);
        load();
      } else {
        const res = await authFetch<{
          data: { raw_key?: string };
        }>(`/api/v1/devices`, {
          method: "POST",
          body: JSON.stringify(toPayload(form, false)),
        });
        setRawKey(res.data?.raw_key ?? "");
        if (!res.data?.raw_key) {
          setShowForm(false);
          load();
        }
      }
    } catch (err) {
      if (!handleAuthError(err))
        setFormError(err instanceof Error ? err.message : "Gagal menyimpan");
    }
  }

  async function handleDelete(dev: Device) {
    if (!confirm(`Hapus device ${dev.id}?`)) return;
    try {
      await authFetch(`/api/v1/devices/${dev.id}`, { method: "DELETE" });
      load();
    } catch (err) {
      if (!handleAuthError(err))
        setError(err instanceof Error ? err.message : "Gagal menghapus");
    }
  }

  async function openDetail(dev: Device) {
    setDetail(dev);
    setHistory([]);
    setCreds([]);
    try {
      const [h, c] = await Promise.all([
        authFetch<{ data: StatusHist[] }>(
          `/api/v1/devices/${dev.id}/health?limit=50`,
        ),
        authFetch<{ data: ApiKeyMeta[] }>(
          `/api/v1/devices/${dev.id}/credentials`,
        ),
      ]);
      setHistory(h.data ?? []);
      setCreds(c.data ?? []);
    } catch (err) {
      if (!handleAuthError(err))
        setError(
          err instanceof Error ? err.message : "Gagal memuat detail device",
        );
    }
  }

  function closeDetail() {
    setDetail(null);
    setHistory([]);
    setCreds([]);
  }

  async function handleRotate(dev: Device) {
    if (!confirm(`Generate API key baru untuk ${dev.id}? Key lama di-revoke.`))
      return;
    setRotated(null);
    try {
      const res = await authFetch<{
        data: { raw_key: string };
      }>(`/api/v1/devices/${dev.id}/credentials/rotate`, { method: "POST" });
      setRotated({ id: dev.id, rawKey: res.data.raw_key });
      if (detail?.id === dev.id) openDetail(dev);
    } catch (err) {
      if (!handleAuthError(err))
        setError(err instanceof Error ? err.message : "Gagal generate key");
    }
  }

  return (
    <div>
      <div className="mb-4 flex items-center justify-between">
        <h1 className="text-2xl font-semibold">Manajemen Device</h1>
        <button
          onClick={openCreate}
          className="rounded bg-zinc-900 px-4 py-2 text-sm font-medium text-white dark:bg-zinc-50 dark:text-zinc-900"
        >
          Tambah Device
        </button>
      </div>

      <form
        className="mb-4 flex flex-col gap-2 sm:flex-row"
        onSubmit={(e) => {
          e.preventDefault();
          setPage(1);
          setQ(qInput);
        }}
      >
        <input
          value={qInput}
          onChange={(e) => setQInput(e.target.value)}
          placeholder="Cari nama / lokasi / id..."
          className="w-full rounded border border-zinc-300 px-3 py-2 dark:border-zinc-700 dark:bg-zinc-800"
        />
        <select
          value={status}
          onChange={(e) => {
            setPage(1);
            setStatus(e.target.value);
          }}
          className="rounded border border-zinc-300 px-3 py-2 dark:border-zinc-700 dark:bg-zinc-800"
        >
          <option value="">Semua status</option>
          {STATUSES.map((s) => (
            <option key={s} value={s}>
              {s}
            </option>
          ))}
        </select>
        <button
          type="submit"
          className="rounded bg-zinc-200 px-4 py-2 text-sm dark:bg-zinc-800"
        >
          Cari
        </button>
      </form>

      {error && (
        <p className="mb-4 rounded bg-red-100 px-3 py-2 text-sm text-red-700">
          {error}
        </p>
      )}

      {rotated && (
        <div className="mb-4 rounded bg-green-100 p-3 text-sm text-green-800">
          <p className="font-medium">
            API key baru untuk {rotated.id} (hanya tampil 1x):
          </p>
          <code className="mt-1 block break-all font-mono">
            {rotated.rawKey}
          </code>
          <button
            onClick={() => setRotated(null)}
            className="mt-2 rounded bg-green-700 px-3 py-1 text-white"
          >
            Tutup
          </button>
        </div>
      )}

      {loading ? (
        <p>Memuat...</p>
      ) : (
        <>
          <div className="overflow-x-auto rounded shadow">
          <table className="w-full min-w-[640px] bg-white text-sm dark:bg-zinc-900">
            <thead>
              <tr className="border-b text-left dark:border-zinc-700">
                <th className="px-3 py-2">ID</th>
                <th className="px-3 py-2">Nama</th>
                <th className="px-3 py-2">Status</th>
                <th className="px-3 py-2">Lokasi</th>
                <th className="px-3 py-2">Aksi</th>
              </tr>
            </thead>
            <tbody>
              {(list?.data ?? []).map((d) => (
                <tr
                  key={d.id}
                  className="border-b last:border-0 dark:border-zinc-800"
                >
                  <td className="px-3 py-2 font-mono">{d.id}</td>
                  <td className="px-3 py-2">{d.name}</td>
                  <td className="px-3 py-2">{d.status}</td>
                  <td className="px-3 py-2">{d.location?.name ?? "-"}</td>
                  <td className="px-3 py-2">
                    <div className="flex gap-2">
                      <button
                        onClick={() => openDetail(d)}
                        className="text-blue-600 hover:underline"
                      >
                        Detail
                      </button>
                      <button
                        onClick={() => openEdit(d)}
                        className="text-amber-600 hover:underline"
                      >
                        Edit
                      </button>
                      <button
                        onClick={() => handleRotate(d)}
                        className="text-purple-600 hover:underline"
                      >
                        API Key
                      </button>
                      <button
                        onClick={() => handleDelete(d)}
                        className="text-red-600 hover:underline"
                      >
                        Hapus
                      </button>
                    </div>
                  </td>
                </tr>
              ))}
              {(list?.data ?? []).length === 0 && (
                <tr>
                  <td colSpan={5} className="px-3 py-4 text-center">
                    Tidak ada data
                  </td>
                </tr>
              )}
            </tbody>
          </table>
          </div>

          <div className="mt-4 flex items-center gap-3 text-sm">
            <button
              disabled={page <= 1}
              onClick={() => setPage((p) => p - 1)}
              className="rounded bg-zinc-200 px-3 py-1 disabled:opacity-50 dark:bg-zinc-800"
            >
              Prev
            </button>
            <span>
              Halaman {list?.page ?? page} / {list?.total_page ?? 1} (total{" "}
              {list?.total ?? 0})
            </span>
            <button
              disabled={page >= (list?.total_page ?? 1)}
              onClick={() => setPage((p) => p + 1)}
              className="rounded bg-zinc-200 px-3 py-1 disabled:opacity-50 dark:bg-zinc-800"
            >
              Next
            </button>
          </div>
        </>
      )}

      {showForm && (
        <div className="fixed inset-0 z-50 flex items-end justify-center bg-black/50 sm:items-center sm:p-4">
          <form
            onSubmit={handleSubmit}
            className="max-h-[92vh] w-full max-w-md overflow-auto rounded-t-2xl bg-white p-6 sm:rounded-lg dark:bg-zinc-900"
          >
            <h2 className="mb-4 text-lg font-semibold">
              {editing ? `Edit ${editing.id}` : "Tambah Device"}
            </h2>
            {formError && (
              <p className="mb-4 rounded bg-red-100 px-3 py-2 text-sm text-red-700">
                {formError}
              </p>
            )}
            {rawKey ? (
              <div className="mb-4 rounded bg-green-100 p-3 text-sm text-green-800">
                <p className="font-medium">
                  Device tersimpan. API key (hanya tampil 1x):
                </p>
                <code className="mt-1 block break-all font-mono">{rawKey}</code>
                <button
                  type="button"
                  onClick={() => {
                    setShowForm(false);
                    setRawKey("");
                    load();
                  }}
                  className="mt-3 rounded bg-zinc-900 px-4 py-2 text-white"
                >
                  Tutup
                </button>
              </div>
            ) : (
              <>
                {!editing && (
                  <label className="mb-3 block">
                    <span className="mb-1 block text-sm">ID</span>
                    <input
                      value={form.id}
                      onChange={(e) =>
                        setForm({ ...form, id: e.target.value })
                      }
                      required
                      className="w-full rounded border px-3 py-2 dark:border-zinc-700 dark:bg-zinc-800"
                    />
                  </label>
                )}
                <label className="mb-3 block">
                  <span className="mb-1 block text-sm">Nama</span>
                  <input
                    value={form.name}
                    onChange={(e) =>
                      setForm({ ...form, name: e.target.value })
                    }
                    required
                    className="w-full rounded border px-3 py-2 dark:border-zinc-700 dark:bg-zinc-800"
                  />
                </label>
                <label className="mb-3 block">
                  <span className="mb-1 block text-sm">Deskripsi</span>
                  <input
                    value={form.description}
                    onChange={(e) =>
                      setForm({ ...form, description: e.target.value })
                    }
                    className="w-full rounded border px-3 py-2 dark:border-zinc-700 dark:bg-zinc-800"
                  />
                </label>
                <label className="mb-3 block">
                  <span className="mb-1 block text-sm">Status</span>
                  <select
                    value={form.status}
                    onChange={(e) =>
                      setForm({ ...form, status: e.target.value })
                    }
                    className="w-full rounded border px-3 py-2 dark:border-zinc-700 dark:bg-zinc-800"
                  >
                    {STATUSES.map((s) => (
                      <option key={s} value={s}>
                        {s}
                      </option>
                    ))}
                  </select>
                </label>
                <label className="mb-3 block">
                  <span className="mb-1 block text-sm">Lokasi</span>
                  <select
                    value={form.location_id}
                    onChange={(e) =>
                      setForm({ ...form, location_id: e.target.value })
                    }
                    className="w-full rounded border px-3 py-2 dark:border-zinc-700 dark:bg-zinc-800"
                  >
                    <option value="">- Tanpa lokasi -</option>
                    {locations.map((l) => (
                      <option key={l.id} value={l.id}>
                        {l.name}
                      </option>
                    ))}
                  </select>
                </label>
                <div className="mb-4 grid grid-cols-3 gap-2">
                  {(["longitude", "latitude", "altitude"] as const).map(
                    (k) => (
                      <label key={k} className="block">
                        <span className="mb-1 block text-sm capitalize">
                          {k === "longitude"
                            ? "Longitude"
                            : k === "latitude"
                              ? "Latitude"
                              : "Altitude"}
                        </span>
                        <input
                          type="number"
                          step="any"
                          value={form[k]}
                          onChange={(e) =>
                            setForm({ ...form, [k]: e.target.value })
                          }
                          className="w-full rounded border px-3 py-2 dark:border-zinc-700 dark:bg-zinc-800"
                        />
                      </label>
                    ),
                  )}
                </div>
                <div className="flex justify-end gap-2">
                  <button
                    type="button"
                    onClick={() => setShowForm(false)}
                    className="rounded bg-zinc-200 px-4 py-2 text-sm dark:bg-zinc-800"
                  >
                    Batal
                  </button>
                  <button
                    type="submit"
                    className="rounded bg-zinc-900 px-4 py-2 text-sm text-white dark:bg-zinc-50 dark:text-zinc-900"
                  >
                    Simpan
                  </button>
                </div>
              </>
            )}
          </form>
        </div>
      )}

      {detail && (
        <div className="fixed inset-0 z-50 flex items-end justify-center bg-black/50 sm:items-center sm:p-4">
          <div className="max-h-[92vh] w-full max-w-lg overflow-auto rounded-t-2xl bg-white p-6 sm:rounded-lg dark:bg-zinc-900">
            <h2 className="mb-4 text-lg font-semibold">
              Detail {detail.id}
            </h2>
            <dl className="mb-4 space-y-1 text-sm">
              <div className="flex gap-2">
                <dt className="w-28 shrink-0 font-medium">Nama</dt>
                <dd>{detail.name}</dd>
              </div>
              <div className="flex gap-2">
                <dt className="w-28 shrink-0 font-medium">Status</dt>
                <dd>{detail.status}</dd>
              </div>
              <div className="flex gap-2">
                <dt className="w-28 shrink-0 font-medium">Lokasi</dt>
                <dd>{detail.location?.name ?? "-"}</dd>
              </div>
              <div className="flex gap-2">
                <dt className="w-28 shrink-0 font-medium">Koordinat</dt>
                <dd>
                  {[detail.longitude, detail.latitude, detail.altitude]
                    .map((v) => (v != null ? String(v) : "-"))
                    .join(", ")}
                </dd>
              </div>
              <div className="flex gap-2">
                <dt className="w-28 shrink-0 font-medium">Deskripsi</dt>
                <dd>{detail.description || "-"}</dd>
              </div>
            </dl>
            <h3 className="mb-2 font-medium">Latest Health</h3>
            <pre className="mb-4 overflow-auto rounded bg-zinc-100 p-3 text-xs dark:bg-zinc-800">
              {detail.latest_health != null
                ? JSON.stringify(detail.latest_health, null, 2)
                : "-"}
            </pre>
            <h3 className="mb-2 font-medium">API Keys</h3>
            {creds.length === 0 ? (
              <p className="mb-4 text-sm">Belum ada API key</p>
            ) : (
              <ul className="mb-4 space-y-1 text-sm">
                {creds.map((c) => (
                  <li key={c.id} className="flex justify-between gap-2">
                    <span className="font-mono">{c.masking}</span>
                    <span className="text-zinc-500">
                      {c.is_revoked ? "revoked" : "aktif"} ·{" "}
                      {new Date(c.created_at).toLocaleString()}
                    </span>
                  </li>
                ))}
              </ul>
            )}
            <h3 className="mb-2 font-medium">Histori Status</h3>
            {history.length === 0 ? (
              <p className="mb-4 text-sm">Belum ada histori</p>
            ) : (
              <ul className="mb-4 space-y-1 text-sm">
                {history.map((h) => (
                  <li key={h.id} className="flex justify-between gap-2">
                    <span>{h.status}</span>
                    <span className="text-zinc-500">
                      {new Date(h.created_at).toLocaleString()}
                    </span>
                  </li>
                ))}
              </ul>
            )}
            <div className="flex justify-end">
              <button
                onClick={closeDetail}
                className="rounded bg-zinc-200 px-4 py-2 text-sm dark:bg-zinc-800"
              >
                Tutup
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
