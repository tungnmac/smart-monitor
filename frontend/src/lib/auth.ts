import type { NextAuthOptions } from "next-auth";
import Credentials from "next-auth/providers/credentials";

const username = process.env.AUTH_USERNAME || "admin";
const password = process.env.AUTH_PASSWORD || "changeme";

const BACKEND_URL = process.env.NEXT_PUBLIC_BACKEND_URL || "http://localhost:8080";

export const authOptions: NextAuthOptions = {
  session: { strategy: "jwt" },
  pages: {
    signIn: "/login",
  },
  providers: [
    Credentials({
      name: "Credentials",
      credentials: {
        email: { label: "Email", type: "text" },
        password: { label: "Password", type: "password" },
      },
      authorize: async (creds) => {
        if (!creds?.email || !creds?.password) return null;
        try {
          const res = await fetch(`${BACKEND_URL}/auth/signin`, {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ email: creds.email, password: creds.password }),
          });
          if (!res.ok) return null;
          const data = await res.json();
          const user = data.user as { id: string; email: string; username?: string; role?: string };
          return {
            id: user.id,
            name: user.username || user.email,
            email: user.email,
            role: user.role || "viewer",
            accessToken: data.token,
          } as any;
        } catch (e) {
          return null;
        }
      },
    }),
  ],
  callbacks: {
    async jwt({ token, user }) {
      if (user) {
        token.role = (user as any).role || "viewer";
        token.accessToken = (user as any).accessToken;
      }
      return token;
    },
    async session({ session, token }) {
      if (session.user) {
        session.user.role = (token as any).role || "admin";
        (session as any).accessToken = (token as any).accessToken;
      }
      return session;
    },
  },
};
