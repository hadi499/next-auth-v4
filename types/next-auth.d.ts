import NextAuth from "next-auth";

// 🔥 Menambahkan properti kustom di sesi NextAuth
declare module "next-auth" {
  interface Session {
    user: {
      id: string; // Pastikan sesuai dengan tipe dari backend (UUID atau string)
      userName?: string; // Opsional jika hanya login Google tanpa username
      name: string;
      email: string;
      role: string;
      provider: string; // Menentukan apakah dari Google atau credentials
      accessToken?: string; // Token jika menggunakan API eksternal
    };
  }

  interface User {
    id: string;
    userName?: string;
    name: string;
    email: string;
    role: string;
    provider: string;
    accessToken?: string;
  }

  

}


