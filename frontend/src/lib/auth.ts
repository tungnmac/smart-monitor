import type { NextAuthOptions } from "next-auth";
import Credentials from "next-auth/providers/credentials";

const BACKEND_URL = process.env.NEXT_PUBLIC_BACKEND_URL || "http://localhost:8080";

export const authOptions: NextAuthOptions = {
  secret: process.env.NEXTAUTH_SECRET,
  session: { strategy: "jwt" },
  pages: {
    signIn: "/login",
    error: "/login", // Redirect errors to login page
  },
  providers: [
    Credentials({
      id: "credentials",
      name: "Credentials",
      credentials: {
        email: { label: "Email", type: "email", placeholder: "email@example.com" },
        password: { label: "Password", type: "password" },
      },
      async authorize(credentials) {
        if (!credentials?.email || !credentials?.password) {
          console.log("Missing credentials");
          return null;
        }
        
        try {
          const res = await fetch(`${BACKEND_URL}/auth/signin`, {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ 
              email: credentials.email, 
              password: credentials.password 
            }),
          });
          
          if (!res.ok) {
            console.log("Backend signin failed:", res.status);
            return null;
          }
          
          const data = await res.json();
          
          if (!data.user || !data.token) {
            console.log("Invalid response format");
            return null;
          }
          
          const user = data.user as { id: string; email: string; username?: string; role?: string };
          return {
            id: user.id,
            name: user.username || user.email,
            email: user.email,
            role: user.role || "viewer",
            accessToken: data.token,
          } as any;
        } catch (e) {
          console.error("Authorization error:", e);
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
