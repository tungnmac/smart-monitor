import { withAuth } from "next-auth/middleware";

// Protect dashboard routes using NextAuth; proxy replaces deprecated middleware naming.
export default withAuth({
  pages: {
    signIn: "/login",
  },
});

export const config = {
  matcher: ["/dashboard/:path*"],
};
