import { useState } from 'react';
import { ConversationList } from './ConversationList';
import { ConversationDetail } from './ConversationDetail';
import type { Conversation } from '../../lib/api';

interface MessagesPageProps {
  currentAgentId: string;
}

export function MessagesPage({ currentAgentId }: MessagesPageProps) {
  const [selectedConversation, setSelectedConversation] = useState<Conversation | null>(null);

  // Responsive: show list on mobile when no conversation selected
  const showList = !selectedConversation;

  return (
    <div className="flex h-[calc(100vh-120px)] rounded-lg overflow-hidden" style={{ backgroundColor: '#0F172A', border: '1px solid #334155' }}>
      {/* Conversation List - hidden on mobile when conversation selected */}
      <div
        className={`w-full md:w-80 lg:w-96 border-r ${showList ? 'block' : 'hidden md:block'}`}
        style={{ borderColor: '#334155' }}
      >
        <ConversationList
          selectedId={selectedConversation?.id}
          onSelect={setSelectedConversation}
        />
      </div>

      {/* Conversation Detail */}
      <div className={`flex-1 ${!showList ? 'block' : 'hidden md:block'}`}>
        {selectedConversation ? (
          <ConversationDetail
            conversation={selectedConversation}
            currentAgentId={currentAgentId}
            onBack={() => setSelectedConversation(null)}
          />
        ) : (
          <div className="h-full flex items-center justify-center" style={{ backgroundColor: '#0F172A' }}>
            <div className="text-center">
              <div
                className="w-16 h-16 rounded-full flex items-center justify-center mx-auto mb-4"
                style={{ backgroundColor: '#1E293B' }}
              >
                <svg
                  className="w-8 h-8 text-[#64748B]"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={1.5}
                    d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z"
                  />
                </svg>
              </div>
              <h3 className="text-white font-medium mb-1">Select a conversation</h3>
              <p className="text-[#64748B] text-sm">
                Choose a conversation from the list to start messaging
              </p>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
