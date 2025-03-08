import NextAuth, { NextAuthOptions, User } from "next-auth";
import CredentialsProvider from "next-auth/providers/credentials";
import GoogleProvider from "next-auth/providers/google";

// ✅ Definisi tipe user
interface CustomUser extends User {
  id: string;
  name: string;
  email: string;
  role: string;
  provider: string;
  accessToken?: string;
  password?: string;
}

// ✅ Fungsi untuk mendaftarkan user Google jika belum ada
const createUser = async (user: Partial<CustomUser>) => {
  try {
    const res = await fetch("http://localhost:8080/register", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(user),
    });

    if (!res.ok) {
      const errorResponse = await res.json();
      console.error("Gagal membuat user:", errorResponse);
      throw new Error("Gagal membuat user");
    }

    return await res.json();
  } catch (error) {
    console.error("Error di createUser:", error);
    throw new Error("Error saat register user Google");
  }
};

export const authOptions: NextAuthOptions = {
  providers: [
    // ✅ Provider Google
    GoogleProvider({
      clientId: process.env.GOOGLE_CLIENT_ID!,
      clientSecret: process.env.GOOGLE_CLIENT_SECRET!,
    }),

    // ✅ Provider Credentials (Manual Login)
    CredentialsProvider({
      name: "Credentials",
      credentials: {
        username: { label: "Username", type: "text", placeholder: "jsmith" },
        password: { label: "Password", type: "password" },
      },
      async authorize(credentials) {
        try {
          const res = await fetch("http://localhost:8080/login", {
            method: "POST",
            headers: {
              "Content-Type": "application/json",
            },
            body: JSON.stringify({
              username: credentials?.username,
              password: credentials?.password,
            }),
          });

          const response = await res.json();
          console.log("Response Login Manual:", response);

          if (response?.user) {
            return {
              id: response.user.id,
              name: response.user.name,
              email: response.user.email,
              role: response.user.role,
              accessToken: response.accessToken,
              provider: "credentials",
            } as CustomUser;
          }

          throw new Error("Username atau password salah");
        } catch (error) {
          console.error("Error di authorize:", error);
          throw new Error("Kesalahan internal");
        }
      },
    }),
  ],

  pages: {
    signIn: "/login", // Custom halaman login
  },

  callbacks: {
    async signIn({ user, account }) {
      console.log("User Sign In:", user, "Provider:", account?.provider);

      // Jika login dari Google, cek apakah user sudah ada di database
      if (account?.provider === "google") {
        try {
          // Cek user di database berdasarkan email
          const res = await fetch(
            `http://localhost:8080/user?email=${user.email}`
          );
          const existingUser = await res.json();

          if (!res.ok) {
            throw new Error("Gagal memeriksa user di database");
          }

          // Jika user belum ada, daftarkan
          if (!existingUser) {
            await createUser({
              name: user.name!,
              email: user.email!,
              provider: "google",
              password: "", // Google user tidak memerlukan password
            });
          }
        } catch (error) {
          console.error("Gagal memproses user Google:", error);
          return false; // Gagal mendaftarkan user
        }
      }

      return true; // Lanjutkan proses login
    },

    async jwt({ token, user, account }) {
      if (user) {
        token.id = (user as CustomUser).id;
        token.name = user.name;
        token.email = user.email;
        token.role = (user as CustomUser).role || "user";
        token.provider = account?.provider || "credentials";
        token.accessToken = (user as CustomUser).accessToken;
      }
      return token;
    },

    async session({ session, token }) {
      session.user = {
        id: token.id as string,
        name: token.name as string,
        email: token.email as string,
        role: token.role as string,
        provider: token.provider as string,
        accessToken: token.accessToken as string,
      };
      return session;
    },

    async redirect({ url, baseUrl }) {
      return url === "/api/auth/signin" ? "/" : url;
    },
  },
};

export default NextAuth(authOptions);
