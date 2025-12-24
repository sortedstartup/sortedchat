import { useState, useEffect } from 'react';
import { Button } from '../../components/ui/button';
import { Card } from '../../components/ui/card';
import { getUIConfig } from '../../lib/config';

// Utility function to detect if running in Wails desktop app
const isWailsDesktopApp = (): boolean => {
  if (typeof window === 'undefined') return false;

  // Check if Wails runtime is available
  // The wailsjs/runtime module injects these into window
  return typeof (window as any).runtime !== 'undefined' ||
    typeof (window as any).go !== 'undefined';
};

export function LoginPage() {
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState('');
  const [isDesktop, setIsDesktop] = useState(false);

  // Detect if running in Wails desktop app
  useEffect(() => {
    setIsDesktop(isWailsDesktopApp());
  }, []);

  // Check if we're coming back from OAuth callback with a token
  useEffect(() => {
    const urlParams = new URLSearchParams(window.location.search);
    const errorParam = urlParams.get('error');

    if (errorParam) {
      setError('Authentication failed. Please try again.');
    }
  }, []);

  const handleGoogleLogin = async () => {
    setIsLoading(true);
    setError('');

    try {
      // Get configuration from runtime config
      const config = getUIConfig();
      if (!config) {
        setError('Application configuration not loaded.');
        setIsLoading(false);
        return;
      }

      const clientId = config.GOOGLE_CLIENT_ID;
      const redirectUri = config.GOOGLE_REDIRECT_URL;

      if (!clientId) {
        setError('Google OAuth is not configured.');
        setIsLoading(false);
        return;
      }

      // Build Google OAuth URL directly (same parameters as backend /login)
      const googleOAuthParams = new URLSearchParams({
        client_id: clientId,
        redirect_uri: redirectUri,
        response_type: 'code',
        scope: 'openid email profile',
        access_type: 'offline',
        state: 'state', // In production, this should be a random value for CSRF protection
      });

      const googleOAuthURL = `${config.GOOGLE_OAUTH_URL}?${googleOAuthParams.toString()}`;

      console.log('Redirecting to Google OAuth:', googleOAuthURL);

      // Redirect to Google OAuth
      window.location.href = googleOAuthURL;
    } catch (err) {
      setError('An error occurred during login');
      console.error('Login error:', err);
      setIsLoading(false);
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-slate-50 to-slate-100 py-12 px-4 sm:px-6 lg:px-8">
      <div className="max-w-md w-full space-y-8">
        {/* Logo/Brand Section */}
        <div className="text-center">
          <div className="mx-auto h-12 w-12 bg-gradient-to-br from-blue-600 to-purple-600 rounded-xl flex items-center justify-center mb-6">
            <svg className="h-6 w-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
            </svg>
          </div>
          <h2 className="text-3xl font-bold tracking-tight text-gray-900">
            Welcome to SortedChat
          </h2>
          <p className="mt-2 text-sm text-gray-600">
            {isDesktop ? 'Sign in to continue' : 'Sign in with your Google account to continue'}
          </p>
        </div>

        <Card className="p-8 shadow-xl border-0 bg-white/80 backdrop-blur-sm">
          {error && (
            <div className="mb-6 bg-red-50 border border-red-200 text-red-600 px-4 py-3 rounded-lg text-sm">
              {error}
            </div>
          )}

          <div className="space-y-6">
            <Button
              onClick={handleGoogleLogin}
              disabled={isLoading}
              className="w-full h-12 bg-white hover:bg-gray-50 text-gray-700 border border-gray-300 shadow-sm transition-all duration-200 hover:shadow-md"
              variant="outline"
            >
              {isLoading ? (
                <div className="flex items-center space-x-2">
                  <div className="w-5 h-5 border-2 border-gray-400 border-t-transparent rounded-full animate-spin"></div>
                  <span>Signing in...</span>
                </div>
              ) : (
                <div className="flex items-center space-x-3">
                  <svg className="w-5 h-5" viewBox="0 0 24 24">
                    <path fill="#4285F4" d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z" />
                    <path fill="#34A853" d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z" />
                    <path fill="#FBBC05" d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z" />
                    <path fill="#EA4335" d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z" />
                  </svg>
                  <span className="font-medium">{isDesktop ? 'Sign in' : 'Continue with Google'}</span>
                </div>
              )}
            </Button>

            <div className="relative">
              <div className="absolute inset-0 flex items-center">
                <div className="w-full border-t border-gray-200"></div>
              </div>
              <div className="relative flex justify-center text-sm">
                <span className="px-2 bg-white text-gray-500">Secure OAuth Authentication</span>
              </div>
            </div>

            <div className="text-xs text-gray-500 text-center space-y-1">
              <p>By signing in, you agree to our terms of service.</p>
              <p>Your data is protected with enterprise-grade security.</p>
            </div>
          </div>
        </Card>

        {/* Footer */}
        <div className="text-center text-xs text-gray-400">
          <p>© 2024 SortedChat. All rights reserved.</p>
        </div>
      </div>
    </div>
  );
}