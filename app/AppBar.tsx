import { getServerSession } from "next-auth";
import { authOptions } from "@/pages/api/auth/[...nextauth]";

export default async function AppBar() {
  // Ambil sesi di server
  const session = await getServerSession(authOptions);

  return (
    <div className="bg-gradient-to-b from-cyan-50 to-cyan-200 p-2 flex gap-5">
      <div className="ml-auto flex gap-2">
        {session?.user ? (
          <>
            <p className="text-sky-600">{session.user.name}</p>
            <form action="/api/auth/signout" method="POST">
              <button type="submit" className="text-red-500">
                Sign Out
              </button>
            </form>
          </>
        ) : (
          <form action="/api/auth/signin" method="POST">
            <button type="submit" className="text-green-600">
              Sign In
            </button>
          </form>
        )}
      </div>
    </div>
  );
}
