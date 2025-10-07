import { useState, useEffect } from 'react';
import { SettingServiceClient, IsFirstBootRequest } from '../../proto/chatservice';

export function useFirstBoot() {
  const [isFirstBoot, setIsFirstBoot] = useState<boolean | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  
  useEffect(() => {
    const checkFirstBoot = async () => {
      try {
        const client = new SettingServiceClient(
          import.meta.env.VITE_API_URL || window.location.origin
        );
        const request = new IsFirstBootRequest({});
        
        const response = await client.IsFirstBoot(request, {});
        setIsFirstBoot(response.is_first_boot);
      } catch (err) {
        console.error('Failed to check first boot status:', err);
        setError(err instanceof Error ? err.message : 'Failed to check first boot status');
        // Default to false if we can't check (assume not first boot)
        setIsFirstBoot(false);
      } finally {
        setIsLoading(false);
      }
    };
    
    checkFirstBoot();
  }, []);
  
  return { isFirstBoot, isLoading, error };
}
